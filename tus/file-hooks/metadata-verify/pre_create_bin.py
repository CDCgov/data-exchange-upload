import getopt
import json
import os
import sys
import uuid
import logging
import math
from argparse import Namespace
from dotenv import load_dotenv

from proc_stat_controller import ProcStatController

load_dotenv()

logger = logging.getLogger(__name__)

REQUIRED_METADATA_FIELDS = ['meta_destination_id', 'meta_ext_event']
STAGE_NAME = 'dex-metadata-verify'


class AllowedDestinationEvents:
    def __init__(self, destination_id, ext_events):
        self.destination_id, self.ext_events = destination_id, ext_events


class EventDefinition:
    def __init__(self, name, definition_filename):
        self.name, self.definition_filename = name, definition_filename


class MetaDefinition:
    def __init__(self, schema_version, fields):
        self.schema_version, self.fields = schema_version, fields


class Field:
    def __init__(self, fieldname, allowed_values, required, description):
        self.fieldname, self.allowed_values, self.required, self.description = fieldname, allowed_values, required, description


def get_version_int_from_str(version):
    l = [int(x, 10) for x in version.split('.')]
    l.reverse()
    version = sum(x * (100 ** i) for i, x in enumerate(l))
    return version


def verify_destination_and_event_allowed(dest_id, event_type):
    config_file = os.path.join(os.path.dirname(__file__), 'allowed_destination_and_events.json')

    with open(config_file, 'r') as file:
        definitions = file.read()
        definitions_dict = json.loads(definitions, object_hook=lambda d: Namespace(**d))

        for definition in definitions_dict:
            if definition.destination_id == dest_id:
                # found the destination_id, checking for the ext_event
                for ext_event in definition.ext_events:
                    if ext_event.name == event_type:
                        return ext_event.definition_filename

        # If we got here, we couldn't find a valid combo of dest_id and event_type.
        failure_message = "Not a recognized combination of meta_destination_id (" + dest_id + ") and meta_ext_event (" + event_type + ")"
        raise Exception(failure_message)


def get_schema_def_by_version(available_schema_defs, requested_schema_version, meta_json):
    selected_schema = None

    if requested_schema_version is not None:
        selected_schema = next(
            (schema_def for schema_def in available_schema_defs if
             schema_def.schema_version == requested_schema_version),
            None)

        if selected_schema is None:
            available_schemas = map(lambda schema_def: schema_def.schema_version, available_schema_defs)
            failure_message = 'Requested schema version ' + requested_schema_version + 'not available.  Available ' \
                                                                                       'schema versions: ' + str(
                available_schemas)
            raise Exception(failure_message)
        else:
            return selected_schema

    # Get the oldest schema
    oldest_available_schema_version_int = math.inf

    for schema_def in available_schema_defs:
        schema_version_int = get_version_int_from_str(schema_def.schema_version)

        if schema_version_int < oldest_available_schema_version_int:
            oldest_available_schema_version_int = schema_version_int
            selected_schema = schema_def

    return selected_schema


def check_program_event_metadata(program_event_meta_filename, metadata):
    # lookup remaining metadata fields specific to this meta_destination_id and meta_ext_event
    with open(program_event_meta_filename, 'r') as file:
        metadata_definition = file.read()
        definition_obj = json.loads(metadata_definition, object_hook=lambda d: Namespace(**d))
        check_metadata_against_definition(definition_obj, metadata)


def check_metadata_against_definition(definition_obj, meta_json):
    # check if the schema was provided and if not, default to the oldest schema
    requested_schema_version = None

    if "schema_version" in meta_json:
        requested_schema_version = str(meta_json["schema_version"])

    schema_to_use = get_schema_def_by_version(definition_obj, requested_schema_version, meta_json)

    print("Using schema_version = " + schema_to_use.schema_version)
    missing_metadata_fields = []
    found_validation_error = False
    validation_error_messages = []

    for field_def in schema_to_use.fields:
        if field_def.required != "false" and field_def.fieldname not in meta_json:
            missing_metadata_fields.append(field_def)

        if field_def.fieldname in meta_json:
            field_value = meta_json[field_def.fieldname]

            if field_def.allowed_values is not None and len(
                    field_def.allowed_values) > 0 and field_value not in field_def.allowed_values:
                validation_error_messages.append(field_def.fieldname + ' = ' + field_value + 'is not one of the allowed '
                                                                                           'values: ' + json.dumps(
                    field_def.allowed_values))
                print(field_def.fieldname + " = " + field_value + " is not one of the allowed values: " + json.dumps(
                    field_def.allowed_values))
                found_validation_error = True

    if len(missing_metadata_fields) > 0:
        for field_def in missing_metadata_fields:
            validation_error_messages.append(
                "Missing required metadata '" + field_def.fieldname + "', description = '" + field_def.description + "'")
            print(
                "Missing required metadata '" + field_def.fieldname + "', description = '" + field_def.description + "'")
            found_validation_error = True

    if found_validation_error:
        raise Exception(stringify_error_messages(validation_error_messages))


def get_required_metadata(meta_json):
    missing_metadata_fields = []

    for field in REQUIRED_METADATA_FIELDS:
        if field not in meta_json:
            missing_metadata_fields.append(field)

    if len(missing_metadata_fields) > 0:
        raise Exception('Missing one or more required metadata fields: ' + str(missing_metadata_fields))

    return [
        meta_json['meta_destination_id'],
        meta_json['meta_ext_event'],
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
    ps_api_controller.create_report(upload_id, destination_id, event_type, STAGE_NAME, payload)

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

    for field in meta_json:
        if field in filename_metadata_fields:
            filename = meta_json[field]

    if filename is None:
        raise Exception('No filename provided.')

    return filename


def verify_metadata(dest_id, event_type, meta_json):
    # check if the program/event type is on the list of allowed
    filename = verify_destination_and_event_allowed(dest_id, event_type)
    if filename is not None:
        check_program_event_metadata(filename, meta_json)


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
        upload_id = report_verification_failure([e], dest_id, event_type, meta_json)
        print(json.dumps({
            'upload_id': upload_id,
            'message': e
        }))
        sys.exit(1)


if __name__ == "__main__":
    main(sys.argv[1:])
