import getopt
import sys
import config
import datetime
import json
import ast
import asyncio
import logging

from azure.servicebus.aio import ServiceBusClient
from azure.servicebus import ServiceBusMessage
from azure.servicebus import TransportType

from types import SimpleNamespace

METADATA_VERSION_ONE = "1.0"
METADATA_VERSION_TWO = "2.0"

logger = logging.getLogger("post-receive-bin")
logger.setLevel(logging.DEBUG)

# remove all default handlers
for handler in logger.handlers:
    logger.removeHandler(handler)

# create console handler and set level to debug
console_handle = logging.StreamHandler()
console_handle.setLevel(logging.DEBUG)

# create formatter
formatter = logging.Formatter(
    fmt='[%(name)s] %(asctime)s.%(msecs)03d %(levelname)s: %(message)s',
    datefmt='%Y/%m/%d %H:%M:%S'
)

console_handle.setFormatter(formatter)

# now add new handler to logger
logger.addHandler(console_handle)

service_bus_connection_str = config.queue_settings['service_bus_connection_str']
queue_name = config.queue_settings['queue_name']

async def send_message(message):
    # Create a Service Bus message and send it to the queue
    message = ServiceBusMessage(message)

    async with ServiceBusClient.from_connection_string(
        conn_str=service_bus_connection_str,
        transport_type=TransportType.AmqpOverWebsocket,
        logging_enable=False) as servicebus_client:
        # Get a Queue Sender object to send messages to the queue
        sender = servicebus_client.get_queue_sender(queue_name=queue_name)
        async with sender:
            await sender.send_messages(message)

async def post_receive(tguid, offset, size, metadata_json):
    try:
        logger.info('python version = {0}'.format(sys.version))
        logger.info('metadata_json = {0}'.format(metadata_json))

        metadata = json.loads(metadata_json, object_hook=lambda d: SimpleNamespace(**d))

        filename = None

        if hasattr(metadata, 'filename'):
            filename = metadata.filename
        elif hasattr(metadata, 'meta_ext_filename'):
            filename = metadata.meta_ext_filename
        elif hasattr(metadata, 'original_filename'):
            filename = metadata.original_filename

        if filename is None:
            raise Exception("filename, meta_ext_filename, or original_filename not found in metadata.")

        logger.info('filename = {0}, metadata_version = {1}'.format(filename, metadata_version))

        # convert metadata json string to a dictionary
        metadata_json_dict = ast.literal_eval(metadata_json)

        json_data, metadata_version = get_report_body(metadata, filename, tguid, offset, size, metadata_json_dict)

        logger.info('post_receive_bin: {0}, offset = {1}'.format(datetime.datetime.now(), offset))

        json_string = json.dumps(json_data)

        logger.info('JSON MESSAGE: %s', json_string)

        await send_message(json_string)

    except Exception as e:
        logger.error("POST RECEIVE HOOK - exiting post_receive with error: %s", str(e), exc_info=True)
        sys.exit(1)


def get_report_body(metadata, filename, tguid, offset, size, metadata_json_dict):
    metadata_version = metadata.metadata_config.version

    if metadata_version == METADATA_VERSION_ONE: 
        json_data = {
            "upload_id": tguid,
            "stage_name": "dex-upload",
            "destination_id": meta_destination_id,
            "event_type": meta_ext_event,
            "content_type": "json",
            "content": {
                        "schema_name": "upload",
                        "schema_version": "1.0",
                        "tguid": tguid,
                        "offset": offset,
                        "size": size,
                        "filename": filename,
                        "meta_destination_id": metadata.meta_destination_id,
                        "meta_ext_event": metadata.meta_ext_event,
                        "metadata": metadata_json_dict
            },
            "disposition_type": "replace"
        }
    elif metadata_version == METADATA_VERSION_TWO:
        json_data = {
            "upload_id": tguid,
            "stage_name": "dex-upload",
            "data_stream_id": data_stream_id,
            "data_stream_route": data_stream_route,
            "content_type": "json",
            "content": {
                        "schema_name": "upload",
                        "schema_version": "1.0",
                        "tguid": tguid,
                        "offset": offset,
                        "size": size,
                        "filename": filename,
                        "data_stream_id": metadata.data_stream_id,
                        "data_stream_route": metadata.data_stream_route,
                        "metadata": metadata_json_dict
            },
            "disposition_type": "replace"
        }
        
    return json_data, metadata_version


def main(argv):
    tus_id = ''
    offset = ''
    size = ''
    metadata = ''
    opts, args = getopt.getopt(argv,"hiosm:",["id=","offset=","size=","metadata="])
    for opt, arg in opts:
        if opt == '-h':
            print('post_receive_bin.py -i id -o offset -s size')
            sys.exit()
        elif opt in ("-i", "--id"):
            tus_id = arg
        elif opt in ("-o", "--offset"):
            offset = arg
        elif opt in ("-s", "--size"):
            size = arg
        elif opt in ("-m", "--metadata"):
            metadata = arg
    try:        
        asyncio.run(post_receive(tus_id, int(offset), int(size), metadata))
    except Exception as e:
        logger.error("POST RECEIVE HOOK - exiting main with error: %s", str(e), exc_info=True)
        sys.exit(1)

if __name__ == "__main__":
    main(sys.argv[1:])