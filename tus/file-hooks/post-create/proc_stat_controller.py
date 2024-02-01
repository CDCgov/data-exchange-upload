import requests


def _handle_trace_response(resp_json):
    if 'trace_id' not in resp_json or 'span_id' not in resp_json:
        raise Exception('Invalid PS API response: ' + str(resp_json))


class ProcStatController:
    def __init__(self, url):
        self.url = url

    def create_upload_trace(self, upload_id, destination_id, event_type):
        params = {
            'uploadId': upload_id,
            'destinationId': destination_id,
            'eventType': event_type
        }
        response = requests.post(f'{self.url}/api/trace', params=params)
        response.raise_for_status()

        resp_json = response.json()

        _handle_trace_response(resp_json)

        return resp_json['trace_id'], resp_json['span_id']

    def start_span_for_trace(self, trace_id, parent_span_id, stage_name):
        params = {
            "stageName": stage_name,
        }
        response = requests.put(f'{self.url}/api/trace/startSpan/{trace_id}/{parent_span_id}', params=params)
        response.raise_for_status()

        resp_json = response.json()

        _handle_trace_response(resp_json)

        return resp_json['trace_id'], resp_json['span_id']
