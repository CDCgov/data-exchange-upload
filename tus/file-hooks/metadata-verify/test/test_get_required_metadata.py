import unittest

from pre_create_bin import get_required_metadata

class TestGetRequiredMetadata(unittest.TestCase):

    def test_version_one_with_all_fields(self):
        meta_json = {
            'meta_destination_id': '123',
            'meta_ext_event': 'event1'
        }
        result = get_required_metadata(meta_json, "1.0")
        self.assertEqual(result, ['123', 'event1'])

    def test_version_two_with_all_fields(self):
        meta_json = {
            'data_stream_id': '456',
            'data_stream_route': 'route1'
        }
        result = get_required_metadata(meta_json, "2.0")
        self.assertEqual(result, ['456', 'route1'])

    def test_missing_required_fields(self):
        meta_json = {}
        
        with self.assertRaises(Exception) as context:
            get_required_metadata(meta_json, "1.0")
        
        self.assertEqual(str(context.exception), "Missing one or more required metadata fields: ['meta_destination_id', 'meta_ext_event']")