import sys, getopt
import json
import requests

required_metadata_fields = ["meta_destination_id", "meta_ext_event"]

def create_upload_trace(uploadId, destinationId, eventType):
  url = 'https://apidev.cdc.gov/processingstatus'
  params = {
    "uploadId": uploadId,
    "destinationId": destinationId,
    "eventType": eventType
  }
  response = requests.post(url, params=params)
  response.raise_for_status()

  resp_json = response.json()

  if "trace_id" not in resp_json or "span_id" not in resp_json:
    raise Exception("Invalid PS API response: " + str(resp_json))
  
  return (resp_json["trace_id"], resp_json["span_id"])

def get_required_metadata(metadata_str):
  meta_json = json.loads(metadata_str)
  missing_metadata_fields = []

  for field in required_metadata_fields:
    if not field in meta_json:
      missing_metadata_fields.append(field)

  if len(missing_metadata_fields) > 0:
    raise Exception("Missing one or more required metadata fields: " + str(missing_metadata_fields))

  return [
    meta_json["meta_destination_id"],
    meta_json["meta_ext_event"],
  ]

def main(argv):
  tguid = None
  metadata = None
  opts, args = getopt.getopt(argv, "him:", ["id=", "metadata="])

  for opt, arg in opts:
    if opt == '-h':
      print('post_create_bin.py -m <inputfile>')
      sys.exit()
    elif opt in ("-i", "--id"):
      tguid = arg
    elif opt in ("-m", "--metadata"):
      metadata = arg
  
  if tguid is None:
    raise Exception("No tguid provided")

  # Create upload trace.
  dest, event = get_required_metadata(metadata)
  trace_id, parent_span_id = create_upload_trace(tguid, dest, event)
  print(trace_id)
  print(parent_span_id)

if __name__ == "__main__":
  main(sys.argv[1:])
