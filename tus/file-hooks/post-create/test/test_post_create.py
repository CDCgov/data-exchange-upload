import unittest

from post_create_bin import get_required_metadata

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

if __name__ == "__main__":
  unittest.main()
