import sys, getopt
import json
from argparse import Namespace

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

def checkIfProgramAndEventAllowed(meta_destination_id, meta_ext_event, metadata):
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
        print("Not a recognized combination of meta_destination_id (" + meta_destination_id + ") and meta_ext_event (" + meta_ext_event + ")")
        sys.exit(1)

def getSchemaVersionToUse(definitionObj, requested_schema_version):
    if requested_schema_version != None:
        available_schemas = []
        for fieldDef in definitionObj:
            if requested_schema_version != None and requested_schema_version == fieldDef.schema_version:
                # found requested schema version
                return fieldDef
            available_schemas.append(str(fieldDef.schema_version))
        print("Requested schema version " + requested_schema_version + " not available.  Available schema versions: " + str(available_schemas))
        sys.exit(1)
    # provide the oldest found schema
    oldest_available_schema_version_int = None
    oldest_available_schema = None
    for fieldDef in definitionObj:
        schema_version_int = getVersionIntFromStr(fieldDef.schema_version)
        #print("found schema_version_int = " + str(schema_version_int))
        if oldest_available_schema == None or schema_version_int < oldest_available_schema_version_int:
            oldest_available_schema_version_int = schema_version_int
            oldest_available_schema = fieldDef
    return oldest_available_schema

def checkProgramEventMetadata(program_event_meta_filename, metadata):
    meta_json = json.loads(metadata)
    # lookup remaining metadata fields specific to this meta_destination_id and meta_ext_event
    with open(program_event_meta_filename, 'r') as file:
        metadata_definition = file.read()
        definitionObj = json.loads(metadata_definition, object_hook=lambda d: Namespace(**d))
        # check if the schema was provided and if not, default to the oldest schema
        requested_schema_version = None
        if "schema_version" in meta_json:
            requested_schema_version = str(meta_json["schema_version"])
        schema_to_use = getSchemaVersionToUse(definitionObj, requested_schema_version)
        print("Using schema_version = " + schema_to_use.schema_version)
        missing_metadata_fields = []
        for fieldDef in schema_to_use.fields:
            if fieldDef.required != "false" and not fieldDef.fieldname in meta_json:
                missing_metadata_fields.append(fieldDef)
        if len(missing_metadata_fields) > 0:
            for fieldDef in missing_metadata_fields:
                print("Missing required metadata '" + fieldDef.fieldname + "', description = '" + fieldDef.description + "'")
            print("Provided metadata: " + metadata)
            sys.exit(1)
            
def checkMetadata(metadata):
    meta_json = json.loads(metadata)
    min_required_metadata = ["meta_destination_id", "meta_ext_event"]
    missing_metadata_fields = []
    for meta_field in min_required_metadata:
        if not meta_field in meta_json:
            missing_metadata_fields.append(meta_field)
    if len(missing_metadata_fields) > 0:
        print("Missing one or more required metadata fields: " + str(missing_metadata_fields) + ", metadata = " + metadata)
        sys.exit(1)
    meta_destination_id = meta_json["meta_destination_id"]
    meta_ext_event = meta_json["meta_ext_event"]
    # check if the program/event type is on the list of allowed
    filename = checkIfProgramAndEventAllowed(meta_destination_id, meta_ext_event, metadata)
    if filename != None:
        checkProgramEventMetadata(filename, metadata)
    
def main(argv):
    metadata = ''
    opts, args = getopt.getopt(argv,"hm:",["metadata="])
    for opt, arg in opts:
        if opt == '-h':
            print ('pre-create-bin.py -m <inputfile>')
            sys.exit()
        elif opt in ("-m", "--metadata"):
            metadata = arg
    checkMetadata(metadata)

if __name__ == "__main__":
    main(sys.argv[1:])
