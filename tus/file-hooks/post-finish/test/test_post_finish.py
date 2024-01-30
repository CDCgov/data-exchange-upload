import unittest
from unittest.mock import patch
from proc_stat_controller import ProcStatController
from post_finish_bin import post_finish


class TestPostFinishMethods(unittest.TestCase):
  @patch.object(ProcStatController, 'get_trace_by_upload_id', return_value=("test_trace", "test_span"))
  @patch.object(ProcStatController, 'stop_span_for_trace')
  def test_should_send_create_trace_request(self, get_trace_by_upload_id, stop_span_for_trace_mock):
    get_trace_by_upload_id.return_value = ("test_trace", "test_span")

    post_finish("1234")

    self.assertEqual(get_trace_by_upload_id.call_count, 1)
    self.assertEqual(stop_span_for_trace_mock.call_count, 1)

if __name__ == "__main__":
  unittest.main()
