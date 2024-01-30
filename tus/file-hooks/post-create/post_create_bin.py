import argparse
import json
import os
import logging
import getopt, sys

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
        meta_json['meta_ext_event'],
    ]

def post_create(dest, event, tguid):
    logger.info(f'Creating trace for upload {tguid} with destination {dest} and event {event}')

    ps_api_controller = ProcStatController(os.getenv('PS_API_URL'))
    trace_id, parent_span_id = ps_api_controller.create_upload_trace(tguid, dest, event)
    logger.debug(f'Created trace for upload {tguid} with trace ID {trace_id} and parent span ID {parent_span_id}')

    # Start the upload child span.  Will be stopped in post-finish hook when the upload is complete.
    ps_api_controller.start_span_for_trace(trace_id, parent_span_id, "dex-upload")
    logger.debug(f'Created child span for parent span {parent_span_id} with stage name of dex-upload')

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
    post_create(dest, event, tguid)


if __name__ == '__main__':
    main(sys.argv[1:])
