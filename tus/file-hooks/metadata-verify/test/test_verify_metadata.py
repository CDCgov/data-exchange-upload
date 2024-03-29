import unittest
from unittest.mock import patch, MagicMock
from pre_create_bin import verify_metadata, get_upload_config, check_metadata_against_config, get_filename_from_metadata, verify_filename

class TestVerifyMetadata(unittest.TestCase):

    @patch('pre_create_bin.get_upload_config')
    @patch('pre_create_bin.check_metadata_against_config')
    def test_valid_upload_config(self, mock_check_metadata, mock_get_upload_config):
        dest_id = '123'
        event_type = 'event1'
        meta_json = {
            'filename': 'example.txt',
            'meta_destination_id': '123',
            'meta_ext_event': 'event1'
        }
        mock_get_upload_config.return_value = {
            'metadata_config': {
                'required_fields': ['meta_destination_id', 'meta_ext_event']
            }
        }
        verify_metadata(dest_id, event_type, meta_json)
        
        mock_get_upload_config.assert_called_once_with(dest_id, event_type)
        mock_check_metadata.assert_called_once_with(meta_json, {'required_fields': ['meta_destination_id', 'meta_ext_event']})

    @patch('pre_create_bin.get_upload_config', return_value=None)
    @patch('pre_create_bin.check_metadata_against_config')
    def test_invalid_upload_config(self, mock_check_metadata, mock_get_upload_config):
        dest_id = '123'
        event_type = 'event1'
        meta_json = {
            'meta_destination_id': '123',
            'meta_ext_event': 'event1'
        }

        with self.assertRaises(Exception) as context:
            verify_metadata(dest_id, event_type, meta_json)

        self.assertEqual(str(context.exception), 'No filename provided.')

        mock_get_upload_config.assert_called_once_with(dest_id, event_type)
        mock_check_metadata.assert_not_called()

    @patch('pre_create_bin.get_upload_config')
    @patch('pre_create_bin.verify_filename')
    @patch('pre_create_bin.get_filename_from_metadata', return_value='test_file.jpg')
    def test_verify_metadata_with_filename(self, mock_get_filename, mock_verify_filename, mock_get_upload_config):
        dest_id = '123'
        event_type = 'event1'
        meta_json = {
            'filename': 'test_file.jpg',
            'meta_destination_id': '123',
            'meta_ext_event': 'event1'
        }
        mock_get_upload_config.return_value = {
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
        verify_metadata(dest_id, event_type, meta_json)
        
        mock_get_upload_config.assert_called_once_with(dest_id, event_type)
        mock_verify_filename.assert_called_once_with('test_file.jpg')
