import unittest
from proc_stat_controller import ProcStatController


class TestProcStatController(unittest.TestCase):
    def test_should_retry_get_trace_when_upload_id_invalid(self):
        controller = ProcStatController('http://dummy', .01)

        with self.assertRaises(Exception) as context:
            controller.get_trace_by_upload_id(None)

        self.assertEqual(controller.retry_count, 1)

    def test_should_retry_get_span_when_upload_id_invalid(self):
        controller = ProcStatController('http://dummy', .01)

        with self.assertRaises(Exception) as context:
            controller.get_span_by_upload_id(None, 'dummy')

        self.assertEqual(controller.retry_count, 1)

    def test_should_retry_stop_span_when_input_invalid(self):
        controller = ProcStatController('http://dummy', .01)

        with self.assertRaises(Exception) as context:
            controller.stop_span_for_trace(None, None)

        self.assertEqual(controller.retry_count, 1)