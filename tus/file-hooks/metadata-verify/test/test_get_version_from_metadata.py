import unittest
from pre_create_bin import get_version_str_from_metadata

class TestGetVersionFromMetadata(unittest.TestCase):

    def test_default_version(self):
        meta_json = {}
        result = get_version_str_from_metadata(meta_json)
        self.assertEqual(result, '1.0')

    def test_specified_version(self):
        meta_json = {'version': '2.0'}
        result = get_version_str_from_metadata(meta_json)
        self.assertEqual(result, "2.0")

    def test_unsupported_version(self):
        meta_json = {'version': '3.0'}
        with self.assertRaises(Exception) as context:
            get_version_str_from_metadata(meta_json)
        
        self.assertEqual(str(context.exception), "Unsupported metadata version: 3.0")