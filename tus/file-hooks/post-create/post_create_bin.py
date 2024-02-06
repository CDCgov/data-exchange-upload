import argparse
import json
import os
import logging
import getopt, sys
import ast
from datetime import datetime 

from dotenv import load_dotenv

from proc_stat_controller import ProcStatController

load_dotenv()

logger = logging.getLogger(__name__)

required_metadata_fields = ['meta_destination_id', 'meta_ext_event']

def get_required_metadata(metadata_str):
    meta_json = json.loads(metadata_str)
    missing_metadata_fields = []

    for field in required_metadata_fields:
        if field not in meta_json:
            missing_metadata_fields.append(field)

    if len(missing_metadata_fields) > 0:
        raise Exception('Missing one or more required metadata fields: ' + str(missing_metadata_fields))

    return [
        meta_json['meta_destination_id'],
        meta_json['meta_ext_event']
    ]

def post_create(dest, event, metadata, tguid):
    logger.info(f'Creating trace for upload {tguid} with destination {dest} and event {event}')

    ps_api_controller = ProcStatController(os.getenv('PS_API_URL'))
    trace_id, parent_span_id = ps_api_controller.create_upload_trace(tguid, dest, event)
    logger.debug(f'Created trace for upload {tguid} with trace ID {trace_id} and parent span ID {parent_span_id}')

    create_metadata_verification_span(ps_api_controller, trace_id, parent_span_id, metadata)

    # Start the upload child span.  Will be stopped in post-finish hook when the upload is complete.
    ps_api_controller.start_span_for_trace(trace_id, parent_span_id, "dex-upload")
    logger.debug(f'Created child span for parent span {parent_span_id} with stage name of dex-upload')


def create_metadata_verification_span(ps_api_controller, trace_id, parent_span_id, metadata):

    try:
        # convert metadata json string to a dictionary
        metadata_json_dict = ast.literal_eval(metadata)

        filename_metadata_fields = ['filename', 'original_filename', 'meta_ext_filename']

        filename = None

        for field in filename_metadata_fields:
            if field in metadata_json_dict:
                filename = metadata_json_dict[field]
                break
        
        json_payload = { 
            "schema_version": "0.0.1",
            "schema_name": "dex-metadata-verify",
            "filename": filename,
            "timestamp": datetime.now().isoformat(),
            "metadata": metadata_json_dict,
            "issues": []
        }

        # Start the upload stage metadata verification span
        trace_id, metadata_verify_span_id = ps_api_controller.start_span_for_trace(trace_id, parent_span_id, "metadata-verify", json_payload)
        logger.debug(f'Started child span {metadata_verify_span_id} with stage name metadata-verify of parent span {parent_span_id}')

        # Stop the upload stage metadata verification span
        if metadata_verify_span_id is not None:
            ps_api_controller.stop_span_for_trace(trace_id, metadata_verify_span_id)
            logger.debug(f'Stopped child span {metadata_verify_span_id} with stage name metadata-verify of parent span {parent_span_id} ')

    except Exception as e:
        logger.error(f"An exception occurred during creation of metadata verification span: {e}")
   
def main(argv):
    
    log_level = logging.INFO
    logging.basicConfig(level=log_level)

    opts, args = getopt.getopt(argv,"im:",["id=", "metadata="])

    for opt, arg in opts:
        if opt in ("-i", "--id"):
            tguid = arg
        elif opt in ("-m", "--metadata"):
            metadata = arg

    if tguid is None:
        raise Exception('No tguid provided')

    # Create upload trace.
    dest, event = get_required_metadata(metadata)
    post_create(dest, event, metadata, tguid)


if __name__ == '__main__':
    main(sys.argv[1:])