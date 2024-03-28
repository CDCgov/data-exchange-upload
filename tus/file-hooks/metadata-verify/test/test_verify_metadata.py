import unittest
from unittest.mock import patch
from pre_create_bin import verify_metadata

class TestVerifyMetadata(unittest.TestCase):

    @patch('pre_create_bin.get_upload_config')
    @patch('pre_create_bin.check_metadata_against_config')
    def test_verify_metadata_success(self, mock_check_metadata_against_config, mock_get_upload_config):
        dest_id = 'some_dest_id'
        event_type = 'some_event_type'
        meta_json = {
            'filename': 'example.txt', 
            'version': '1.0'
        }
        
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

        mock_get_upload_config.return_value = {'metadata_config': mock_config}
        
        verify_metadata(dest_id, event_type, meta_json)
        
        mock_get_upload_config.assert_called_once_with(dest_id, event_type, '1')
        mock_check_metadata_against_config.assert_called_once_with(meta_json, mock_config)


    @patch('pre_create_bin.get_upload_config')
    def test_verify_metadata_missing_required_field(self, mock_get_upload_config):
        dest_id = 'some_dest_id'
        event_type = 'some_event_type'
        meta_json = {}
        
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

        mock_get_upload_config.return_value = {'metadata_config': mock_config}

        with self.assertRaises(Exception) as context:
            verify_metadata(dest_id, event_type, meta_json)
        
        self.assertTrue("Missing required metadata 'filename', description = 'The name of the file submitted.'" in str(context.exception))


    @patch('pre_create_bin.get_upload_config')
    def test_verify_metadata_invalid_field_value(self, mock_get_upload_config):
        dest_id = 'some_dest_id'
        event_type = 'some_event_type'
        meta_json = {
            'wrong_field_name': 'example.txt'
        }
        
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

        mock_get_upload_config.return_value = {'metadata_config': mock_config}

        with self.assertRaises(Exception) as context:
            verify_metadata(dest_id, event_type, meta_json)
        
        self.assertTrue("Missing required metadata 'filename', description = 'The name of the file submitted.'" in str(context.exception))

