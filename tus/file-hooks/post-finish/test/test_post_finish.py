import unittest
from unittest.mock import patch
from proc_stat_controller import ProcStatController
from post_finish_bin import post_finish


class TestPostFinishMethods(unittest.TestCase):
    @patch.object(ProcStatController, 'get_span_by_upload_id')
    @patch.object(ProcStatController, 'stop_span_for_trace')
    def test_should_send_requests_to_ps_api(self, get_span_by_upload_id_mock, stop_span_for_trace_mock):
        get_span_by_upload_id_mock.return_value = {
            "trace_id": "1234",
            "span_id": "5678"
        }

        post_finish("1234")

        self.assertEqual(get_span_by_upload_id_mock.call_count, 1)
        self.assertEqual(stop_span_for_trace_mock.call_count, 1)


if __name__ == "__main__":
    unittest.main()
