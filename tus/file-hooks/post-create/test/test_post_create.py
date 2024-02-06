import unittest
import sys, os
from unittest.mock import patch

from post_create_bin import get_required_metadata
from post_create_bin import get_filename_from_metadata
from post_create_bin import post_create

from proc_stat_controller import ProcStatController


class TestPostCreateMethods(unittest.TestCase):
  def test_get_required_metadata(self):
    # Test data
    metadata_json_dict = {'meta_destination_id': '123', 'meta_ext_event': '456'}

    # Call the function to be tested
    result = get_required_metadata(metadata_json_dict)

    # Assert the result
    self.assertEqual(result, ['123', '456'])

  def test_should_raise_error_if_required_metadata_not_provided(self):
    # Test data with missing required fields
    metadata_json_dict = {'other_field': 'value'}

    # Call the function and assert that it raises an exception
    with self.assertRaises(Exception) as context:
      get_required_metadata(metadata_json_dict)

    # Assert the exception message
    self.assertTrue('Missing one or more required metadata fields' in str(context.exception))

  def test_get_filename_from_metadata(self):
    # Test data
    metadata_json_dict = {'filename': 'example.txt', 'original_filename': 'example_original.txt'}

    # Call the function to be tested
    result = get_filename_from_metadata(metadata_json_dict)

    # Assert the result
    self.assertEqual(result, 'example.txt')
  
  @patch.object(ProcStatController, 'create_upload_trace', return_value=("trace_id", "parent_span_id"))
  @patch.object(ProcStatController, 'start_span_for_trace', return_value=("trace_id", "metadata_verify_span_id"))
  @patch.object(ProcStatController, 'create_report_json')
  @patch.object(ProcStatController, 'stop_span_for_trace')
  def test_should_send_create_trace_request(self, create_upload_trace_mock, start_span_for_trace_mock, create_report_json_mock, stop_span_for_trace_mock):
    
    # Test data
    dest = 'destination'
    event = 'some_event'
    metadata_json_dict = {'meta_destination_id': 'destination', 'meta_ext_event': 'some_event'}
    tguid = 'some_tguid'
    
    # Call the function
    post_create(dest, event, metadata_json_dict, tguid)

    # Assert the result
    self.assertEqual(create_upload_trace_mock.call_count, 1)
    self.assertEqual(start_span_for_trace_mock.call_count, 1)
    self.assertEqual(create_report_json_mock.call_count, 1)
    self.assertEqual(stop_span_for_trace_mock.call_count, 1)

if __name__ == "__main__":
  unittest.main()
