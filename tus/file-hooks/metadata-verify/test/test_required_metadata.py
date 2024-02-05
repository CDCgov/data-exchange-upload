import unittest
from unittest.mock import patch

from pre_create_bin import verify_metadata
from proc_stat_controller import ProcStatController


@patch.object(ProcStatController, 'create_upload_trace', return_value=('test_trace', 'test_span'))
@patch.object(ProcStatController, 'start_span_for_trace', return_value=('test_trace', 'test_child_span_id'))
@patch.object(ProcStatController, 'stop_span_for_trace')
class TestRequiredMetadata(unittest.TestCase):
    def test_missing_metadata_destination_id(self, create_upload_trace_mock, start_span_for_trace_mock,
                                             stop_span_for_trace_mock):
        metadata = """
        {
            "meta_ext_event":"456"
        }
        """
        with self.assertRaises(Exception) as context:
            verify_metadata(metadata)

        self.assertIn('Missing one or more required metadata fields', str(context.exception))
        self.assertEqual(create_upload_trace_mock.call_count, 1)
        self.assertEqual(start_span_for_trace_mock.call_count, 1)
        self.assertEqual(stop_span_for_trace_mock.call_count, 1)

    def test_missing_metadata_ext_event(self, create_upload_trace_mock, start_span_for_trace_mock,
                                        stop_span_for_trace_mock):
        metadata = """
        {
            "meta_destination_id":"1234"
        }
        """
        with self.assertRaises(Exception) as context:
            verify_metadata(metadata)

        self.assertIn('Missing one or more required metadata fields', str(context.exception))
        self.assertEqual(create_upload_trace_mock.call_count, 1)
        self.assertEqual(start_span_for_trace_mock.call_count, 1)
        self.assertEqual(stop_span_for_trace_mock.call_count, 1)


if __name__ == '__main__':
    unittest.main()
