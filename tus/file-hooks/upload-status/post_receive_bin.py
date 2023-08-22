import getopt
import sys
import config
import datetime
import base64

from azure.storage.queue import QueueServiceClient
from azure.core.exceptions import ResourceExistsError

import json
import ast
import logging

from types import SimpleNamespace

# Include secondary dependencies here, since pyinstaller will miss them and
# the binary will fail at run-time.
import chardet

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

account_name = config.queue_settings['storage_account_name']
account_key = config.queue_settings['storage_account_key']
q_name = config.queue_settings['queue_name']

def get_queue_client():
    connect_str = 'DefaultEndpointsProtocol=https;AccountName={0};AccountKey={1};EndpointSuffix=core.windows.net'.format(account_name, account_key)
    service_client = QueueServiceClient.from_connection_string(connect_str)
    try:
        return service_client.create_queue(q_name)
    except ResourceExistsError:
        # Queue exists.  Note, you will get a false positive that the resource still exists if the
        # queue was very recently deleted.  There seems to be a couple minute delay where after a
        # queue is deleted that Azure still reports the resource exists.
        pass
    return service_client.get_queue_client(q_name)

def send_message_to_cosmos_sync(json_update):
    try:
        logger.debug('send_message_to_cosmos_sync: sending update message to queue: {0}'.format(json_update))
        json_update_base64_bytes = base64.b64encode(bytes(json_update, 'utf-8')) # bytes
        base64_str = json_update_base64_bytes.decode('utf-8') # convert bytes to string
        get_queue_client().send_message(base64_str)
    except Exception as e:
        logger.exception(e)

def upsert_item(tguid, offset, size, filename, meta_destination_id, meta_ext_event, metadata_json):
    logger.info('Upserting tguid = {0}'.format(tguid))

    logger.info('tguid: {0}'.format(tguid))
    logger.info('offset: {0}'.format(offset))
    logger.info('size: {0}'.format(size))

    logger.info('Sending update to queue: {0}'.format(q_name))
    update = {
        'tguid': tguid,
        'offset': offset,
        'size': size,
        'filename': filename,
        'meta_destination_id': meta_destination_id,
        'meta_ext_event': meta_ext_event,
        'metadata': metadata_json
    }
    json_update = json.dumps(update)
    send_message_to_cosmos_sync(json_update)

def post_receive(tguid, offset, size, metadata_json):
    try:
        logger.info('python version = {0}'.format(sys.version))
        logger.info('metadata_json = {0}'.format(metadata_json))

        metadata = json.loads(metadata_json, object_hook=lambda d: SimpleNamespace(**d))

        filename = None

        if metadata.filename != None:
            filename = metadata.filename
        elif metadata.meta_ext_filename != None:
            filename = metadata.meta_ext_filename
        elif metadata.original_filename != None:
            filename = metadata.original_filename

        if filename is None:
            raise Exception("filename, meta_ext_filename, or original_filename not found in metadata.")

        meta_destination_id = metadata.meta_destination_id
        meta_ext_event = metadata.meta_ext_event
        logger.info('filename = {0}, meta_destination_id = {1}, meta_ext_event = {2}'.format(filename, meta_destination_id, meta_ext_event))

        # convert metadata json string to a dictionary
        metadata_json_dict = ast.literal_eval(metadata_json)

        logger.info('post_receive_bin: {0}, offset = {1}'.format(datetime.datetime.now(), offset))
        upsert_item(tguid, offset, size, filename, meta_destination_id, meta_ext_event, metadata_json_dict)
    except Exception as e:
        print(e)
        sys.exit(1)

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
        post_receive(tus_id, int(offset), int(size), metadata)
    except Exception as e:
        print(e)
        sys.exit(1)

if __name__ == "__main__":
    main(sys.argv[1:])