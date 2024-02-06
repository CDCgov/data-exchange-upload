import getopt
import json
import os
import sys
import uuid
import logging
import math
from argparse import Namespace

from proc_stat_controller import ProcStatController

logger = logging.getLogger(__name__)

required_metadata_fields = ['meta_destination_id', 'meta_ext_event']


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


def getVersionIntFromStr(version):
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
        # handle_verification_failure([failure_message], dest_id, event_type)


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
            # dest_id, event = get_required_metadata(meta_json)
            # handle_verification_failure([failure_message], dest_id, event)
        else:
            return selected_schema

    # Get the oldest schema
    oldest_available_schema_version_int = math.inf

    for schema_def in available_schema_defs:
        schema_version_int = getVersionIntFromStr(schema_def.schema_version)

        if schema_version_int < oldest_available_schema_version_int:
            oldest_available_schema_version_int = schema_version_int
            selected_schema = schema_def

    return selected_schema


def checkProgramEventMetadata(program_event_meta_filename, metadata):
    # lookup remaining metadata fields specific to this meta_destination_id and meta_ext_event
    with open(program_event_meta_filename, 'r') as file:
        metadata_definition = file.read()
        definitionObj = json.loads(metadata_definition, object_hook=lambda d: Namespace(**d))
        checkMetadataAgainstDefinition(definitionObj, metadata)


def checkMetadataAgainstDefinition(definitionObj, meta_json):
    # check if the schema was provided and if not, default to the oldest schema
    requested_schema_version = None

    if "schema_version" in meta_json:
        requested_schema_version = str(meta_json["schema_version"])

    schema_to_use = get_schema_def_by_version(definitionObj, requested_schema_version, meta_json)

    print("Using schema_version = " + schema_to_use.schema_version)
    missing_metadata_fields = []
    validationError = False
    validation_error_messages = []

    for fieldDef in schema_to_use.fields:
        if fieldDef.required != "false" and not fieldDef.fieldname in meta_json:
            missing_metadata_fields.append(fieldDef)

        if fieldDef.fieldname in meta_json:
            fieldValue = meta_json[fieldDef.fieldname]

            if fieldDef.allowed_values != None and len(
                    fieldDef.allowed_values) > 0 and fieldValue not in fieldDef.allowed_values:
                validation_error_messages.append(fieldDef.fieldname + ' = ' + fieldValue + 'is not one of the allowed '
                                                                                           'values: ' + json.dumps(
                    fieldDef.allowed_values))
                print(fieldDef.fieldname + " = " + fieldValue + " is not one of the allowed values: " + json.dumps(
                    fieldDef.allowed_values))
                validationError = True

    if len(missing_metadata_fields) > 0:
        for fieldDef in missing_metadata_fields:
            validation_error_messages.append(
                "Missing required metadata '" + fieldDef.fieldname + "', description = '" + fieldDef.description + "'")
            print(
                "Missing required metadata '" + fieldDef.fieldname + "', description = '" + fieldDef.description + "'")
            validationError = True

    if validationError:
        # dest_id, event = get_required_metadata(meta_json)
        # handle_verification_failure(validation_error_messages, dest_id, event)
        raise Exception(stringify_error_messages(validation_error_messages))


def get_required_metadata(meta_json):
    missing_metadata_fields = []

    for field in required_metadata_fields:
        if field not in meta_json:
            missing_metadata_fields.append(field)

    if len(missing_metadata_fields) > 0:
        # failure_message = 'Missing one or more required metadata fields: ' + str(missing_metadata_fields)
        raise Exception('Missing one or more required metadata fields: ' + str(missing_metadata_fields))

        # dest_id = "not provided"
        # if 'meta_destination_id' not in missing_metadata_fields:
        #     dest_id = meta_json['meta_destination_id']
        #
        # event_type = "not provided"
        # if 'meta_ext_event' not in missing_metadata_fields:
        #     event_type = meta_json['meta_ext_event']
        #
        # handle_verification_failure([failure_message], dest_id, event_type)

    return [
        meta_json['meta_destination_id'],
        meta_json['meta_ext_event'],
    ]


def handle_verification_failure(messages, destination_id, event_type, meta_json):
    ps_api_controller = ProcStatController(os.getenv('PS_API_URL'))

    # Create trace for upload
    upload_id = uuid.uuid4()
    trace_id, parent_span_id = ps_api_controller.create_upload_trace(upload_id, destination_id, event_type)

    # Start the upload stage metadata verification span
    trace_id, metadata_verify_span_id \
        = ps_api_controller.start_span_for_trace(trace_id, parent_span_id, 'metadata-verify')
    logger.debug(
        f'Started child span {metadata_verify_span_id} with stage name metadata-verify of parent span {parent_span_id}')

    filename = get_filename_from_metadata(meta_json)
    # Send report with metadata failure issues.
    payload = {
        'schema_version': '0.0.1',
        'schema_name': 'metadata-verify',
        'filename': filename,
        'metadata': meta_json,
        'issues': messages
    }
    ps_api_controller.create_report(upload_id, destination_id, event_type, payload)

    # Stop the upload stage metadata verification span
    ps_api_controller.stop_span_for_trace(trace_id, metadata_verify_span_id)
    logger.debug(
        f'Stopped child span {metadata_verify_span_id} with stage name metadata-verify of parent span {parent_span_id} ')

    raise Exception(stringify_error_messages(messages))


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
        checkProgramEventMetadata(filename, meta_json)


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
        handle_verification_failure([e], dest_id, event_type, meta_json)
        print(e)
        sys.exit(1)


if __name__ == "__main__":
    main(sys.argv[1:])
