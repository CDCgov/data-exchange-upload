import unittest
from unittest.mock import patch

from post_create_bin import get_required_metadata
from post_create_bin import get_filename_from_metadata
from post_create_bin import post_create

from proc_stat_controller import ProcStatController

class TestPostCreateMethods(unittest.TestCase):
    
    def test_get_required_metadata_version_1(self):
        metadata_json_dict = {
            'version': '1.0',
            'meta_destination_id': '123',
            'meta_ext_event': '456'
        }

        result = get_required_metadata(metadata_json_dict)

        self.assertEqual(result, ['123', '456'])

    def test_get_required_metadata_version_2(self):
        metadata_json_dict = {
            'version': '2.0',
            'data_stream_id': 'stream_id_123',
            'data_stream_route': 'route_123'
        }

        result = get_required_metadata(metadata_json_dict)

        self.assertEqual(result, ['stream_id_123', 'route_123'])

    def test_should_raise_error_if_required_metadata_not_provided(self):
      metadata_json_dict = {
        'version': '2.0'
      }

      with self.assertRaises(Exception) as context:
        get_required_metadata(metadata_json_dict)

      self.assertIn('Missing one or more required metadata fields: ', str(context.exception))

    def test_get_filename_from_metadata(self):
        metadata_json_dict = {'filename': 'example.txt', 'original_filename': 'example_original.txt'}

        result = get_filename_from_metadata(metadata_json_dict)

        self.assertEqual(result, 'example.txt')
  
    @patch('post_create_bin.processing_status_traces_enabled', True)
    @patch.object(ProcStatController, 'create_upload_trace', return_value=("trace_id", "parent_span_id"))
    @patch.object(ProcStatController, 'start_span_for_trace', return_value=("trace_id", "metadata_verify_span_id"))
    @patch.object(ProcStatController, 'create_report_json')
    @patch.object(ProcStatController, 'stop_span_for_trace')
    def test_should_send_create_trace_request(
        self,
        stop_span_for_trace_mock,
        create_report_json_mock,
        start_span_for_trace_mock,
        create_upload_trace_mock
        ):
        from post_create_bin import post_create
        use_case = 'use_case'
        use_case_category = 'some_use_case_category'
        metadata_json_dict = {'meta_destination_id': 'destination', 'meta_ext_event': 'some_event'}
        tguid = 'some_tguid'
        
        post_create(use_case, use_case_category, metadata_json_dict, tguid)

        self.assertEqual(create_upload_trace_mock.call_count, 1)
        self.assertEqual(start_span_for_trace_mock.call_count, 2)
        self.assertEqual(stop_span_for_trace_mock.call_count, 1)

if __name__ == "__main__":
    unittest.main()