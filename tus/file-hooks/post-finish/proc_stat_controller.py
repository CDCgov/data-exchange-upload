import requests


def _handle_trace_response(resp_json):
    if 'trace_id' not in resp_json or 'span_id' not in resp_json:
        raise Exception('Invalid PS API response: ' + str(resp_json))


def _handle_span_response(resp_json):
    if 'trace_id' not in resp_json or 'span_id' not in resp_json:
        raise Exception('Invalid PS API response: ' + str(resp_json))


class ProcStatController:
    def __init__(self, url):
        self.url = url

    def get_trace_by_upload_id(self, upload_id):
        response = requests.get(f'{self.url}/api/trace/uploadId/{upload_id}')
        response.raise_for_status()

        resp_json = response.json()
        _handle_trace_response(resp_json)

        return resp_json['trace_id'], resp_json['span_id']

    def get_span_by_upload_id(self, upload_id, stage_name):
        params = {
            "uploadId": upload_id,
            "stageName": stage_name
        }
        response = requests.get(f'{self.url}/api/trace/span', params=params)
        response.raise_for_status()

        resp_json = response.json()
        _handle_span_response(resp_json)

        return resp_json

    def stop_span_for_trace(self, trace_id, parent_span_id):
        response = requests.put(f'{self.url}/api/trace/stopSpan/{trace_id}/{parent_span_id}')
        response.raise_for_status()
