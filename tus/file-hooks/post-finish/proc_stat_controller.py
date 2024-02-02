import requests
import os
import logging
import time
from requests import Request, Session

MAX_RETRIES = os.getenv("PS_API_MAX_RETRIES") or 6


def _handle_trace_response(resp_json):
    if 'trace_id' not in resp_json or 'span_id' not in resp_json:
        raise Exception('Invalid PS API response: ' + str(resp_json))


def _handle_span_response(resp_json):
    if 'trace_id' not in resp_json or 'span_id' not in resp_json:
        raise Exception('Invalid PS API response: ' + str(resp_json))


class ProcStatController:
    def __init__(self, url, delay_s=1):
        self.url = url
        self.delay_s = delay_s
        self.session = Session()
        self.retry_count = 0
        self.logger = logging.getLogger(__name__)

    def __del__(self):
        self.session.close()

    def get_trace_by_upload_id(self, upload_id):
        req = Request('GET', f'{self.url}/api/trace/uploadId/{upload_id}')
        response = self._send_request_with_retry(req.prepare())

        resp_json = response.json()
        _handle_trace_response(resp_json)

        return resp_json['trace_id'], resp_json['span_id']

    def get_span_by_upload_id(self, upload_id, stage_name):
        params = {
            "uploadId": upload_id,
            "stageName": stage_name
        }
        req = Request('GET', f'{self.url}/api/trace/span', params=params)
        response = self._send_request_with_retry(req.prepare())

        resp_json = response.json()
        _handle_span_response(resp_json)

        return resp_json

    def stop_span_for_trace(self, trace_id, parent_span_id):
        req = Request('PUT', f'{self.url}/api/trace/stopSpan/{trace_id}/{parent_span_id}')
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
