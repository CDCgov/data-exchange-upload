import unittest
import sys, os
from unittest.mock import patch

from post_create_bin import get_required_metadata
from post_create_bin import post_create

sys.path.append(os.path.join(sys.path[0], '..', 'common'))
from proc_stat_controller import ProcStatController

class TestPostCreateMethods(unittest.TestCase):
  def test_should_get_required_metadata(self):
    valid_metadata_str = """
      {
        "meta_ext_event": "test_event",
        "meta_destination_id": "1234"
      }
    """

    dest, event = get_required_metadata(valid_metadata_str)
    self.assertEqual(dest, "1234")
    self.assertEqual(event, "test_event")
  
  def test_should_raise_error_if_required_metadata_not_provided(self):
    metadata = """
      {
          "meta_ext_event":"456"
      }
    """
    with self.assertRaises(Exception) as context:
      get_required_metadata(metadata)
      
    self.assertIn('Missing one or more required metadata fields', str(context.exception))

  @patch.object(ProcStatController, 'create_upload_trace', return_value=("test_trace", "test_span"))
  @patch.object(ProcStatController, 'start_span_for_trace')
  def test_should_send_create_trace_request(self, create_upload_trace_mock, start_span_for_trace_mock):
    create_upload_trace_mock.return_value = ("test_trace", "test_span")

    post_create("1234", "test_event", "5678")

    self.assertEqual(create_upload_trace_mock.call_count, 1)
    self.assertEqual(start_span_for_trace_mock.call_count, 1)

if __name__ == "__main__":
  unittest.main()
