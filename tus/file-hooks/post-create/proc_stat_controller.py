import requests
import os
import json

MAX_RETRIES = os.getenv("PS_API_MAX_RETRIES") or 6

def _handle_trace_response(resp_json):
    if 'trace_id' not in resp_json or 'span_id' not in resp_json:
        raise Exception('Invalid PS API response: ' + str(resp_json))

class ProcStatController:
    def __init__(self, url, delay_s=1):
        self.url = url
        self.delay_s = delay_s
        self.session = Session()
        self.retry_count = 0
        self.logger = logging.getLogger(__name__)

    def __init__(self, url):
        self.url = url

    def create_upload_trace(self, upload_id, destination_id, event_type):
        params = {
            'uploadId': upload_id,
            'destinationId': destination_id,
            'eventType': event_type
        }

        request = Request('POST', f'{self.url}/api/trace', params=params)
        response = self._send_request_with_retry(request.prepare())

        resp_json = response.json()

        _handle_trace_response(resp_json)

        return resp_json['trace_id'], resp_json['span_id']

    def start_span_for_trace(self, trace_id, parent_span_id, stage_name):
        params = {
            "stageName": stage_name,
        }
        
        request = Request('PUT', f'{self.url}/api/trace/stopSpan/{trace_id}/{span_id}', params=params)
        response = self._send_request_with_retry(request.prepare())

        resp_json = response.json()

        _handle_trace_response(resp_json)

        return resp_json['trace_id'], resp_json['span_id']

    def create_report_json(self, upload_id, destination_id, event_type, json_payload):        
        params = {
            'destinationId': destination_id,
            'eventType': event_type
        }

        request = Request('PUT', f'{self.url}/report/json/uploadId/{upload_id}', params=params, json=json_payload)
        self._send_request_with_retry(request.prepare())
        
    def stop_span_for_trace(self, trace_id, span_id):
        req = Request('PUT', f'{self.url}/api/trace/stopSpan/{trace_id}/{span_id}')
        self._send_request_with_retry(req.prepare())

    def _send_request_with_retry(self, req):
        # Resetting the retry count.
        self.retry_count = 0

        while self.retry_count < MAX_RETRIES:
            try:
                resp = self.session.send(req)
                resp.raise_for_status()

                if resp.ok:
                    # Request was handled successfully, return and don't send any more requests.
                    return resp
            except requests.exceptions.RequestException as e:
                self.logger.warning(f"Error sending request to PS API after attempt {self.retry_count}.  Reason: {e}")
                self.retry_count = self.retry_count + 1

                # Waiting 2 second before trying again.
                time.sleep(self.delay_s)

        raise Exception(f"Unable to send successful request to PS API after {MAX_RETRIES} attempts.")