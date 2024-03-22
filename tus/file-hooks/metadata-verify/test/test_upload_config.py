import unittest
from unittest.mock import Mock, patch
from pre_create_bin import get_upload_config

class TestUploadConfig(unittest.TestCase):
    
    @patch('pre_create_bin.DEX_STORAGE_ACCOUNT_SERVICE')
    def test_get_upload_config_success(self, mock_blob_service_client):
        mock_blob_client = Mock()
        mock_blob_service_client.get_blob_client.return_value = mock_blob_client
        mock_blob_client.exists.return_value = True
        mock_blob_client.download_blob.return_value.readall.return_value = b'{"example_key": "example_value"}'

        dest_id = 'destination_id'
        event_type = 'event_type'

        upload_config = get_upload_config(dest_id, event_type)

        self.assertEqual(upload_config, {"example_key": "example_value"})

    @patch('pre_create_bin.DEX_STORAGE_ACCOUNT_SERVICE')
    def test_get_upload_config_blob_not_exists(self, mock_blob_service_client):
        mock_blob_client = Mock()
        mock_blob_service_client.get_blob_client.return_value = mock_blob_client
        mock_blob_client.exists.return_value = False

        dest_id = 'destination_id'
        event_type = 'event_type'

        with self.assertRaises(Exception) as context:
            get_upload_config(dest_id, event_type)

        self.assertEqual(str(context.exception), "Failed to read upload config file provided")

    @patch('pre_create_bin.DEX_STORAGE_ACCOUNT_SERVICE')
    def test_get_upload_config_exception(self, mock_blob_service_client):
        mock_blob_client = Mock()
        mock_blob_service_client.get_blob_client.return_value = mock_blob_client
        mock_blob_client.exists.side_effect = Exception("Mocked exception")

        dest_id = 'destination_id'
        event_type = 'event_type'

        with self.assertRaises(Exception) as context:
            get_upload_config(dest_id, event_type)

        self.assertEqual(str(context.exception), "Failed to read upload config file provided")
