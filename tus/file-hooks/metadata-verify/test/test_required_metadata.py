import unittest

from pre_create_bin import get_required_metadata


class TestRequiredMetadata(unittest.TestCase):
    def test_missing_metadata_destination_id(self):
        metadata = {
            'meta_ext_event': '456'
        }

        with self.assertRaises(Exception) as context:
            get_required_metadata(metadata)

        self.assertIn('Missing one or more required metadata fields', str(context.exception))

    def test_missing_metadata_ext_event(self):
        metadata = {
            'meta_destination_id': '1234'
        }
        with self.assertRaises(Exception) as context:
            get_required_metadata(metadata)

        self.assertIn('Missing one or more required metadata fields', str(context.exception))


if __name__ == '__main__':
    unittest.main()