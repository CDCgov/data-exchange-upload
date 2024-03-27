import unittest
from proc_stat_controller import ProcStatController
from httmock import all_requests, HTTMock


class TestProcStatController(unittest.TestCase):
    def test_should_retry_get_trace_when_upload_id_invalid(self):
        controller = ProcStatController('http://dummy', .01)

        with self.assertRaises(Exception) as context:
            controller.get_trace_by_upload_id(None)

        self.assertEqual(controller.retry_count, 0)

    def test_should_retry_get_span_when_upload_id_invalid(self):
        controller = ProcStatController('http://dummy', .01)

        with self.assertRaises(Exception) as context:
            controller.get_span_by_upload_id(None, 'dummy')

        self.assertEqual(controller.retry_count, 0)

    def test_should_retry_stop_span_when_input_invalid(self):
        controller = ProcStatController('http://dummy', .01)

        with self.assertRaises(Exception) as context:
            controller.stop_span_for_trace(None, None)

        self.assertEqual(controller.retry_count, 0)

    def test_should_retry_429(self):
        controller = ProcStatController('http://dummy', .01)

        with HTTMock(response_retry_limit):
            with self.assertRaises(Exception) as context:
                controller.stop_span_for_trace(None, None)

        self.assertEqual(controller.retry_count, 6)

    def test_should_retry_429_with_after(self):
        controller = ProcStatController('http://dummy', .01)

        with HTTMock(response_retry_limit_with_after):
            with self.assertRaises(Exception) as context:
                controller.stop_span_for_trace(None, None)

        self.assertEqual(controller.retry_count, 6)

    def test_should_retry_503(self):
        controller = ProcStatController('http://dummy', .01)

        with HTTMock(response_retry_unavailable):
            with self.assertRaises(Exception) as context:
                controller.stop_span_for_trace(None, None)

        self.assertEqual(controller.retry_count, 6)

    def test_should_retry_503_with_after(self):
        controller = ProcStatController('http://dummy', .01)

        with HTTMock(response_retry_unavailable_with_after):
            with self.assertRaises(Exception) as context:
                controller.stop_span_for_trace(None, None)

        self.assertEqual(controller.retry_count, 6)

    def test_should_succeed(self):
        controller = ProcStatController('http://dummy', .01)

        with HTTMock(response_success):
            controller.stop_span_for_trace(None, None)

        self.assertEqual(controller.retry_count, 0)


@all_requests
def response_success(url, request):
    return {'status_code': 200}

@all_requests
def response_retry_limit(url, request):
    return {'status_code': 429}

@all_requests
def response_retry_limit_with_after(url, request):
    return {
        'status_code': 429,
        'headers': {'Retry-After': 1}
    }

@all_requests
def response_retry_unavailable(url, request):
    return {'status_code': 503}

@all_requests
def response_retry_unavailable_with_after(url, request):
    return {
        'status_code': 503,
        'headers': {'Retry-After': 1}
    }