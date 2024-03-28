import getopt
import json
import os
import sys
import uuid
import logging
from argparse import Namespace
from dotenv import load_dotenv

from azure.storage.blob import BlobServiceClient

from proc_stat_controller import ProcStatController

load_dotenv()

logger = logging.getLogger(__name__)

METADATA_VERSION_ONE = "1.0"
METADATA_VERSION_TWO = "2.0"
REQUIRED_VERSION_ONE_FIELDS = ['meta_destination_id', 'meta_ext_event']
REQUIRED_VERSION_TWO_FIELDS = ['data_stream_id', 'data_stream_route']

STAGE_NAME = 'dex-metadata-verify'

AZURE_STORAGE_ACCOUNT = os.getenv('AZURE_STORAGE_ACCOUNT')
AZURE_STORAGE_KEY = os.getenv('AZURE_STORAGE_KEY')
UPLOAD_CONFIG_CONTAINER = os.getenv('UPLOAD_CONFIG_CONTAINER')

CONNECTION_STRING = f"DefaultEndpointsProtocol=https;AccountName={AZURE_STORAGE_ACCOUNT};AccountKey={AZURE_STORAGE_KEY};EndpointSuffix=core.windows.net"
DEX_STORAGE_ACCOUNT_SERVICE = BlobServiceClient.from_connection_string(conn_str=CONNECTION_STRING)


def get_upload_config(dest_id, event_type, metadata_version):
    if dest_id is None or event_type is None:
        raise Exception("dest_id and event_type are required in metadata")

    try:
        upload_config_file = f"v{metadata_version}/{dest_id}-{event_type}.json"
        blob_client = DEX_STORAGE_ACCOUNT_SERVICE.get_blob_client(container=UPLOAD_CONFIG_CONTAINER, blob=upload_config_file)

        if not blob_client.exists():
            failure_message = "Not a recognized combination of meta_destination_id (" + dest_id + ") and meta_ext_event (" + event_type + ")"
            raise Exception(failure_message)
        
        downloader = blob_client.download_blob(max_concurrency=1, encoding='UTF-8')
        blob_text = downloader.readall()
        upload_config_data = json.loads(blob_text) 

        return upload_config_data
    
    except Exception as e:
        failure_message = "Failed to read upload config file provided"
        raise Exception(failure_message) from e
    

def check_metadata_against_config(meta_json, meta_config):
    missing_metadata_fields = []
    found_validation_error = False
    validation_error_messages = []

    for field in meta_config['fields']:
        if field['required'] == True and field['field_name'] not in meta_json:
            missing_metadata_fields.append(field)
        
        if field['field_name'] in meta_json:
            field_value = meta_json[field['field_name']]

            if field['allowed_values'] is not None and len(
                    field['allowed_values']) > 0 and field_value not in field['allowed_values']:
                validation_error_messages.append(field['field_name'] + ' = ' + field_value + 'is not one of the allowed '
                                                                                           'values: ' + json.dumps(
                    field['allowed_values']))
                print(field['field_name'] + " = " + field_value + " is not one of the allowed values: " + json.dumps(
                    field['allowed_values']))
                found_validation_error = True

    if len(missing_metadata_fields) > 0:
        for field_def in missing_metadata_fields:
            validation_error_messages.append(
                "Missing required metadata '" + field_def['field_name'] + "', description = '" + field_def['description'] + "'")
            print(
                "Missing required metadata '" + field_def['field_name'] + "', description = '" + field_def['description'] + "'")
            found_validation_error = True

    if found_validation_error:
        raise Exception(stringify_error_messages(validation_error_messages))


def get_required_metadata(meta_json):
    metadata_version = meta_json.get('version')

    if metadata_version == METADATA_VERSION_TWO:
        required_fields = REQUIRED_VERSION_TWO_FIELDS
    elif metadata_version == METADATA_VERSION_ONE:
        required_fields = REQUIRED_VERSION_ONE_FIELDS
    else:
        raise Exception(f"Unsupported metadata version: {metadata_version}")

    missing_metadata_fields = [field for field in required_fields if field not in meta_json]

    if len(missing_metadata_fields) > 0:
        raise Exception('Missing one or more required metadata fields: ' + str(missing_metadata_fields))

    if metadata_version == METADATA_VERSION_TWO:
        return [
            meta_json['data_stream_id'],
            meta_json['data_stream_route']
        ]
    else:
        return [
            meta_json['meta_destination_id'],
            meta_json['meta_ext_event']
        ]


def report_verification_failure(messages, destination_id, event_type, meta_json):
    if destination_id is None:
        destination_id = 'NOT_PROVIDED'

    if event_type is None:
        event_type = 'NOT_PROVIDED'

    ps_api_controller = ProcStatController(os.getenv('PS_API_URL'))

    # Create trace for upload
    upload_id = uuid.uuid4()
    trace_id, parent_span_id = ps_api_controller.create_upload_trace(upload_id, destination_id, event_type)

    # Start the upload stage metadata verification span
    trace_id, metadata_verify_span_id \
        = ps_api_controller.start_span_for_trace(trace_id, parent_span_id, STAGE_NAME)
    logger.debug(
        f'Started child span {metadata_verify_span_id} with stage name metadata-verify of parent span {parent_span_id}')

    filename = get_filename_from_metadata(meta_json)
    # Send report with metadata failure issues.
    payload = {
        'schema_version': '0.0.1',
        'schema_name': STAGE_NAME,
        'filename': filename,
        'metadata': meta_json,
        'issues': messages
    }
    
    ps_api_controller.create_report(upload_id, destination_id, event_type, STAGE_NAME, json.dumps(payload))

    # Stop the upload stage metadata verification span
    ps_api_controller.stop_span_for_trace(trace_id, metadata_verify_span_id)
    logger.debug(
        f'Stopped child span {metadata_verify_span_id} with stage name metadata-verify of parent span {parent_span_id} ')

    return upload_id


def stringify_error_messages(messages):
    return 'Found the following metadata validation errors: ' + ','.join(messages)


def get_filename_from_metadata(meta_json):
    filename_metadata_fields = ['filename', 'original_filename', 'meta_ext_filename']
    filename = None

    for field in filename_metadata_fields:
        if field in meta_json:
            filename = meta_json[field]
            break

    if filename is None:
        raise Exception('No filename provided.')

    return filename


def verify_metadata(dest_id, event_type, meta_json):
    metadata_version = meta_json.get('version')
    metadata_version = metadata_version.split('.')[0] if metadata_version else None

    # check if the program/event type is on the list of allowed
    upload_config = get_upload_config(dest_id, event_type, metadata_version)

    if upload_config is not None:
        check_metadata_against_config(meta_json, upload_config['metadata_config'])


def main(argv):
    log_level = logging.INFO
    logging.basicConfig(level=log_level)

    metadata = ''
    opts, args = getopt.getopt(argv, "hm:", ["metadata="])
    for opt, arg in opts:
        if opt == '-h':
            print('pre-create-bin.py -m <inputfile>')
            sys.exit()
        elif opt in ("-m", "--metadata"):
            metadata = arg

    meta_json = None
    dest_id = None
    event_type = None

    try:
        meta_json = json.loads(metadata)
        dest_id, event_type = get_required_metadata(meta_json)
        verify_metadata(dest_id, event_type, meta_json)
    except Exception as e:
        error_msg = str(e)
        upload_id = report_verification_failure([error_msg], dest_id, event_type, meta_json)
        print(json.dumps({
            'upload_id': str(upload_id),
            'message': error_msg
        }))
        sys.exit(1)


if __name__ == "__main__":
    main(sys.argv[1:])

