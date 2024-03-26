import unittest

from pre_create_bin import get_required_metadata

class TestRequiredMetadata(unittest.TestCase):
    def test_missing_metadata_destination_id_version_1(self):
        metadata = {
            'metadata_config': {
                'version': '1.0'
            },
            'meta_ext_event': '456'
        }

        with self.assertRaises(Exception) as context:
            get_required_metadata(metadata)

        self.assertIn('Missing one or more required metadata fields', str(context.exception))

    def test_missing_metadata_ext_event_version_1(self):
        metadata = {
            'metadata_config': {
                'version': '1.0'
            },
            'meta_destination_id': '1234'
        }

        with self.assertRaises(Exception) as context:
            get_required_metadata(metadata)

        self.assertIn('Missing one or more required metadata fields', str(context.exception))

    def test_missing_metadata_data_stream_id_version_2(self):
        metadata = {
            'metadata_config': {
                'version': '2.0'
            },
            'data_stream_route': 'ndlp'
        }

        with self.assertRaises(Exception) as context:
            get_required_metadata(metadata)

        self.assertIn('Missing one or more required metadata fields', str(context.exception))

    def test_missing_metadata_data_stream_route_version_2(self):
        metadata = {
            'metadata_config': {
                'version': '2.0'
            },
            'data_stream_id': 'routineImmunization'
        }

        with self.assertRaises(Exception) as context:
            get_required_metadata(metadata)

        self.assertIn('Missing one or more required metadata fields', str(context.exception))

if __name__ == '__main__':
    unittest.main()