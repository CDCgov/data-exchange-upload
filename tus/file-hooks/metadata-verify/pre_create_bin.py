import getopt
import json
import os
import sys
import uuid
import logging
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


def checkIfProgramAndEventAllowed(meta_destination_id, meta_ext_event):
    allowed_programs_and_events_filename = "allowed_destination_and_events.json"
    with open(allowed_programs_and_events_filename, 'r') as file:
        definitions = file.read()
        definitionsObj = json.loads(definitions, object_hook=lambda d: Namespace(**d))
        for definition in definitionsObj:
            if definition.destination_id == meta_destination_id:
                # found the destination_id, checking for the ext_event
                for ext_event in definition.ext_events:
                    if ext_event.name == meta_ext_event:
                        return ext_event.definition_filename
        raise Exception(
            "Not a recognized combination of meta_destination_id (" + meta_destination_id + ") and meta_ext_event (" + meta_ext_event + ")")


def getSchemaVersionToUse(definitionObj, requested_schema_version):
    if requested_schema_version != None:
        available_schemas = []
        for fieldDef in definitionObj:
            if requested_schema_version != None and requested_schema_version == fieldDef.schema_version:
                # found requested schema version
                return fieldDef
            available_schemas.append(str(fieldDef.schema_version))
        raise Exception(
            "Requested schema version " + requested_schema_version + " not available.  Available schema versions: " + str(
                available_schemas))
    # provide the oldest found schema
    oldest_available_schema_version_int = None
    oldest_available_schema = None
    for fieldDef in definitionObj:
        schema_version_int = getVersionIntFromStr(fieldDef.schema_version)
        # print("found schema_version_int = " + str(schema_version_int))
        if oldest_available_schema == None or schema_version_int < oldest_available_schema_version_int:
            oldest_available_schema_version_int = schema_version_int
            oldest_available_schema = fieldDef
    return oldest_available_schema


def checkProgramEventMetadata(program_event_meta_filename, metadata):
    # lookup remaining metadata fields specific to this meta_destination_id and meta_ext_event
    with open(program_event_meta_filename, 'r') as file:
        metadata_definition = file.read()
        definitionObj = json.loads(metadata_definition, object_hook=lambda d: Namespace(**d))
        checkMetadataAgainstDefinition(definitionObj, metadata)


def checkMetadataAgainstDefinition(definitionObj, metadata):
    # check if the schema was provided and if not, default to the oldest schema
    meta_json = json.loads(metadata)
    requested_schema_version = None
    if "schema_version" in meta_json:
        requested_schema_version = str(meta_json["schema_version"])
    schema_to_use = getSchemaVersionToUse(definitionObj, requested_schema_version)
    print("Using schema_version = " + schema_to_use.schema_version)
    missing_metadata_fields = []
    validationError = False
    for fieldDef in schema_to_use.fields:
        if fieldDef.required != "false" and not fieldDef.fieldname in meta_json:
            missing_metadata_fields.append(fieldDef)
        if fieldDef.fieldname in meta_json:
            fieldValue = meta_json[fieldDef.fieldname]
            if fieldDef.allowed_values != None and len(
                    fieldDef.allowed_values) > 0 and fieldValue not in fieldDef.allowed_values:
                print(fieldDef.fieldname + " = " + fieldValue + " is not one of the allowed values: " + json.dumps(
                    fieldDef.allowed_values))
                validationError = True
    if len(missing_metadata_fields) > 0:
        for fieldDef in missing_metadata_fields:
            print(
                "Missing required metadata '" + fieldDef.fieldname + "', description = '" + fieldDef.description + "'")
            validationError = True
    if validationError:
        raise Exception("Provided metadata: " + metadata)


def get_required_metadata(metadata_str):
    meta_json = json.loads(metadata_str)
    missing_metadata_fields = []

    for field in required_metadata_fields:
        if field not in meta_json:
            missing_metadata_fields.append(field)

    if len(missing_metadata_fields) > 0:
        failure_message = 'Missing one or more required metadata fields: ' + str(missing_metadata_fields)

        dest_id = "not provided"
        if 'meta_destination_id' not in missing_metadata_fields:
            dest_id = meta_json['meta_destination_id']

        event_type = "not provided"
        if 'meta_ext_event' not in missing_metadata_fields:
            event_type = meta_json['meta_ext_event']

        handle_verification_failure(failure_message, dest_id, event_type)

    return [
        meta_json['meta_destination_id'],
        meta_json['meta_ext_event'],
    ]


def handle_verification_failure(message, destination_id, event_type):
    ps_api_controller = ProcStatController(os.getenv('PS_API_URL'))

    # Create trace for upload
    upload_id = uuid.uuid4()
    trace_id, parent_span_id = ps_api_controller.create_upload_trace(upload_id, destination_id, event_type)

    # Start the upload stage metadata verification span
    trace_id, metadata_verify_span_id \
        = ps_api_controller.start_span_for_trace(trace_id, parent_span_id, 'metadata-verify')
    logger.debug(
        f'Started child span {metadata_verify_span_id} with stage name metadata-verify of parent span {parent_span_id}')

    # TODO: Send report with metadata failure issues.

    # Stop the upload stage metadata verification span
    ps_api_controller.stop_span_for_trace(trace_id, metadata_verify_span_id)
    logger.debug(
        f'Stopped child span {metadata_verify_span_id} with stage name metadata-verify of parent span {parent_span_id} ')

    raise Exception(message)


def verify_metadata(metadata):
    dest_id, event = get_required_metadata(metadata)

    # check if the program/event type is on the list of allowed
    filename = checkIfProgramAndEventAllowed(dest_id, event)
    if filename is not None:
        checkProgramEventMetadata(filename, metadata)


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
    try:
        verify_metadata(metadata)
    except Exception as e:
        print(e)
        sys.exit(1)


if __name__ == "__main__":
    main(sys.argv[1:])
