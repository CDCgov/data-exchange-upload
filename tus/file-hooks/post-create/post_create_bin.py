import sys, argparse, os, time
import json
import requests
from dotenv import load_dotenv

load_dotenv()

required_metadata_fields = ['meta_destination_id', 'meta_ext_event']

def create_upload_trace(uploadId, destinationId, eventType):
  params = {
    'uploadId': uploadId,
    'destinationId': destinationId,
    'eventType': eventType
  }
  response = requests.post(f'{os.getenv("PS_API_URL")}/api/trace', params=params)
  response.raise_for_status()

  resp_json = response.json()

  if 'trace_id' not in resp_json or 'span_id' not in resp_json:
    raise Exception('Invalid PS API response: ' + str(resp_json))
  
  return (resp_json['trace_id'], resp_json['span_id'])

def start_span_for_trace(trace_id, parent_span_id, stage_name):
  params = {
    "stageName": stage_name,
    "spanMark": "start"
  }
  response = requests.put(f'{os.getenv("PS_API_URL")}/api/trace/addSpan/{trace_id}/{parent_span_id}', params=params)
  response.raise_for_status()

  resp_json = response.json()

  if 'trace_id' not in resp_json or 'span_id' not in resp_json:
    raise Exception('Invalid PS API response: ' + str(resp_json))
  
  return (resp_json['trace_id'], resp_json['span_id'])

def get_required_metadata(metadata_str):
  meta_json = json.loads(metadata_str)
  missing_metadata_fields = []

  for field in required_metadata_fields:
    if not field in meta_json:
      missing_metadata_fields.append(field)

  if len(missing_metadata_fields) > 0:
    raise Exception('Missing one or more required metadata fields: ' + str(missing_metadata_fields))

  return [
    meta_json['meta_destination_id'],
    meta_json['meta_ext_event'],
  ]

def main(argv):
  tguid = None
  metadata = None

  parser = argparse.ArgumentParser()
  parser.add_argument('-i', '--id')
  parser.add_argument('-m', '--metadata')

  args = parser.parse_args()
  tguid = args.id
  metadata = args.metadata
  
  if tguid is None:
    raise Exception('No tguid provided')

  # Create upload trace.
  dest, event = get_required_metadata(metadata)
  trace_id, parent_span_id = create_upload_trace(tguid, dest, event)

  # Start the upload child span.  Will be stopped in post-finish hook when the upload is complete.
  start_span_for_trace(trace_id, parent_span_id, "dex-upload")
  

if __name__ == '__main__':
  main(sys.argv[1:])
