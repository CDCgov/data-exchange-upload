import unittest
from unittest.mock import patch
from pre_create_bin import verify_metadata

class TestVerifyMetadata(unittest.TestCase):

    @patch('pre_create_bin.get_upload_config')
    @patch('pre_create_bin.check_metadata_against_config')
    def test_valid_upload_config(self, mock_check_metadata, mock_get_upload_config):
        use_case = '123'
        use_case_category = 'event1'
        version_num = 1
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
        verify_metadata(use_case, use_case_category, meta_json, version_num)
        
        mock_get_upload_config.assert_called_once_with(use_case, use_case_category, version_num)
        mock_check_metadata.assert_called_once_with(meta_json, {'required_fields': ['meta_destination_id', 'meta_ext_event']})

    @patch('pre_create_bin.get_upload_config', return_value=None)
    @patch('pre_create_bin.check_metadata_against_config')
    def test_invalid_upload_config(self, mock_check_metadata, mock_get_upload_config):
        use_case = '123'
        use_case_category = 'event1'
        version_num = 1
        meta_json = {
            'meta_destination_id': '123',
            'meta_ext_event': 'event1'
        }

        with self.assertRaises(Exception) as context:
            verify_metadata(use_case, use_case_category, meta_json, version_num)

        self.assertEqual(str(context.exception), 'No filename provided.')

        mock_get_upload_config.assert_called_once_with(use_case, use_case_category, version_num)
        mock_check_metadata.assert_not_called()

    @patch('pre_create_bin.verify_filename')
    @patch('pre_create_bin.get_upload_config')
    def test_verify_metadata_with_filename(self, mock_get_upload_config, mock_verify_filename):
        use_case = '123'
        use_case_category = 'event1'
        version_num = 1
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
        verify_metadata(use_case, use_case_category, meta_json, version_num)
        
        mock_get_upload_config.assert_called_once_with(use_case, use_case_category, version_num)
        mock_verify_filename.assert_called_once_with('test_file.jpg')
