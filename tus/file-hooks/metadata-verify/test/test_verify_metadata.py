import unittest
from unittest.mock import patch
from pre_create_bin import verify_metadata, get_upload_config

class TestVerifyMetadata(unittest.TestCase):

    @patch('pre_create_bin.get_upload_config')
    def test_verify_metadata_success(self, mock_get_upload_config):
        dest_id = 'some_dest_id'
        event_type = 'some_event_type'
        meta_json = {'filename': 'example.txt'}
        
        mock_config = {
            'metadata_config': {
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
        }

        mock_get_upload_config.return_value = mock_config

        verify_metadata(dest_id, event_type, meta_json)

    @patch('pre_create_bin.get_upload_config')
    def test_verify_metadata_missing_required_field(self, mock_get_upload_config):
        dest_id = 'some_dest_id'
        event_type = 'some_event_type'
        meta_json = {}
        
        mock_config = {
            'metadata_config': {
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
        }

        mock_get_upload_config.return_value = mock_config

        with self.assertRaises(Exception) as context:
            verify_metadata(dest_id, event_type, meta_json)
        
        self.assertTrue("Missing required metadata 'filename', description = 'The name of the file submitted.'" in str(context.exception))

