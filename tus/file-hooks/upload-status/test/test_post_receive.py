import unittest
from unittest.mock import patch, MagicMock, AsyncMock
from post_receive_bin import post_receive
import json
from types import SimpleNamespace

class TestPostReceive(unittest.TestCase):

    @patch('post_receive_bin.logger')
    @patch('post_receive_bin.json.loads')
    @patch('post_receive_bin.ast.literal_eval')
    @patch('post_receive_bin.send_message')
    async def test_post_receive_v2_metadata(self, mock_send_message, mock_ast_eval, mock_json_loads, mock_logger):
        
        tguid = "123456"
        offset = 0
        size = 100
        metadata_json = json.dumps({
            "filename": "test_file.txt",
            "meta_destination_id": "dest_123",
            "meta_ext_event": "event_123",
            "metadata_config": {"version": "2.0"},
            "data_stream_id": "stream_123",
            "data_stream_route": "route_123"
        })

        mock_json_loads.return_value = SimpleNamespace(filename="test_file.txt", meta_destination_id="dest_123",
                                                       meta_ext_event="event_123", metadata_config=SimpleNamespace(version="2.0"),
                                                       data_stream_id="stream_123", data_stream_route="route_123")

        await post_receive(tguid, offset, size, metadata_json)

        mock_logger.info.assert_called_with('filename = test_file.txt, metadata_version = 2.0')
        
        mock_send_message.assert_called()

    @patch('post_receive_bin.logger')
    @patch('post_receive_bin.json.loads')
    @patch('post_receive_bin.ast.literal_eval')
    @patch('post_receive_bin.send_message')
    async def test_post_receive_v1_metadata(self, mock_send_message, mock_ast_eval, mock_json_loads, mock_logger):
        
        tguid = "123456"
        offset = 0
        size = 100
        metadata_json = json.dumps({
            "meta_destination_id": "dest_123",
            "meta_ext_event": "event_123",
            "metadata_config": {"version": "1.0"}
        })

        mock_json_loads.return_value = SimpleNamespace(meta_destination_id="dest_123",
                                                       meta_ext_event="event_123", metadata_config=SimpleNamespace(version="1.0"))

        await post_receive(tguid, offset, size, metadata_json)

        mock_logger.info.assert_called_with('filename = None, metadata_version = 1.0')
        
        mock_send_message.assert_called()

if __name__ == '__main__':
    unittest.main()