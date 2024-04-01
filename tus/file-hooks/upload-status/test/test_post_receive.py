import unittest
import json
from unittest.mock import patch, Mock
from post_receive_bin import post_receive, send_message
import asyncio

class TestPostReceive(unittest.TestCase):

    def setUp(self):
        self.tguid = "123456"
        self.offset = "100"
        self.size = "200"
        self.metadata_json = {
            "filename": "test.txt",
            "meta_destination_id": "dest_id",
            "meta_ext_event": "event"
        }

        self.expected_json_data = {
            "upload_id": self.tguid,
            "stage_name": "dex-upload",
            "destination_id": "dest_id",
            "event_type": "event",
            "content_type": "json",
            "content": {
                "schema_name": "upload",
                "schema_version": "1.0",
                "tguid": self.tguid,
                "offset": int(self.offset),
                "size": int(self.size),
                "filename": "test.txt",
                "meta_destination_id": "dest_id",
                "meta_ext_event": "event",
                "metadata": {}
            },
            "disposition_type": "replace"
        }

        self.mock_get_report_body = Mock(return_value=self.expected_json_data)
        self.mock_sender = Mock()
        self.mock_sender.send_messages.return_value = None
        self.mock_get_queue_sender = Mock(return_value=self.mock_sender)
        self.mock_service_bus_client = Mock()
        self.mock_service_bus_client.get_queue_sender = self.mock_get_queue_sender
        self.mock_from_connection_string = Mock(return_value=self.mock_service_bus_client)

    @patch('post_receive_bin.get_report_body')
    @patch('post_receive_bin.send_message')
    async def test_post_receive_success(self, mock_send_message, mock_get_report_body):
        mock_get_report_body.return_value = self.expected_json_data
        
        post_receive(self.tguid, int(self.offset), int(self.size), json.dumps(self.metadata_json))

        mock_get_report_body.assert_called_once()

    @patch('post_receive_bin.get_report_body')
    @patch('post_receive_bin.send_message')
    async def test_post_receive_exception(self, mock_send_message, mock_get_report_body):
        mock_get_report_body.return_value = self.expected_json_data

        tguid = "123456"
        offset = "100"
        size = "200"
        metadata_json = '{"invalid_key": "test.txt"}'

        with self.assertRaises(Exception):
            await post_receive(self.tguid, int(self.offset), int(self.size), json.dumps(metadata_json))

if __name__ == '__main__':
    unittest.main()