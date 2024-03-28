import unittest
from unittest.mock import patch
from pre_create_bin import check_metadata_against_config

class TestVerifyMetadata(unittest.TestCase):

    @patch('pre_create_bin.get_upload_config')
    def test_check_metadata_against_config_success(self, mock_get_upload_config):
        mock_config = {
            'version': '1.0',
            'fields': [
                {
                    'field_name': 'filename',
                    'allowed_values': None,
                    'required': True,
                    'description': 'The name of the file submitted.'
                }
            ]
        }
        mock_get_upload_config.return_value = mock_config

        meta_json = {
            'filename': 'example.txt', 
            'version': '1.0'
        }

        try:
            check_metadata_against_config(meta_json, mock_config)
        except Exception as e:
            self.fail(f"check_metadata_against_config raised Exception unexpectedly: {e}")

    @patch('pre_create_bin.get_upload_config')
    def test_check_metadata_against_config_missing_required_field(self, mock_get_upload_config):
        mock_config = {
            'version': '1.0',
            'fields': [
                {
                    'field_name': 'filename',
                    'allowed_values': None,
                    'required': True,
                    'description': 'The name of the file submitted.'
                }
            ]
        }
        mock_get_upload_config.return_value = mock_config

        meta_json = { 
            'version': '1.0'
        }

        with self.assertRaises(Exception) as context:
            check_metadata_against_config(meta_json, mock_config)

        self.assertTrue("Missing required metadata 'filename'" in str(context.exception))
