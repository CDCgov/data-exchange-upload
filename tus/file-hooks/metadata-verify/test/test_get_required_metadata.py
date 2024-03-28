import unittest

from pre_create_bin import get_required_metadata

class TestGetRequiredMetadata(unittest.TestCase):

    def test_valid_metadata_version_one(self):
        meta_json = {
            'version': "1.0",
            'meta_destination_id': '123',
            'meta_ext_event': 'route_one'
        }
        result = get_required_metadata(meta_json)
        self.assertEqual(result, ['123', 'route_one'])

    def test_valid_metadata_version_two(self):
        meta_json = {
            'version': "2.0",
            'data_stream_id': '456',
            'data_stream_route': 'aims-celr'
        }
        result = get_required_metadata(meta_json)
        self.assertEqual(result, ['456', 'aims-celr'])

    def test_unsupported_metadata_version(self):
        meta_json = {
            'version': "3.0",
            'data_stream_id': '123',
            'data_stream_route': 'route_one'
        }
        with self.assertRaises(Exception) as context:
            get_required_metadata(meta_json)
        self.assertEqual(str(context.exception), "Unsupported metadata version: 3.0")
