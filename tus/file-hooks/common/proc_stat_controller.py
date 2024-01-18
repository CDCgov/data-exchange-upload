import requests

class ProcStatController:
    def __init__(self, url):
        self.url = url

    def create_upload_trace(self, uploadId, destinationId, eventType):
        params = {
            'uploadId': uploadId,
            'destinationId': destinationId,
            'eventType': eventType
        }
        response = requests.post(f'{self.url}/api/trace', params=params)
        response.raise_for_status()

        resp_json = response.json()

        self._handle_trace_response(resp_json)
        
        return (resp_json['trace_id'], resp_json['span_id'])
    
    def start_span_for_trace(self, trace_id, parent_span_id, stage_name):
        params = {
            "stageName": stage_name,
            "spanMark": "start"
        }
        response = requests.put(f'{self.url}/api/trace/addSpan/{trace_id}/{parent_span_id}', params=params)
        response.raise_for_status()

        resp_json = response.json()

        self._handle_trace_response(resp_json)
        
        return (resp_json['trace_id'], resp_json['span_id'])
    
    def _handle_trace_response(resp_json):
        if 'trace_id' not in resp_json or 'span_id' not in resp_json:
            raise Exception('Invalid PS API response: ' + str(resp_json))