import os, sys
import time
import azure.cosmos.cosmos_client as cosmos_client
import azure.cosmos.exceptions as exceptions
from azure.cosmos.partition_key import PartitionKey

import config
import datetime
import zmq

import logging
from logging.handlers import RotatingFileHandler

import json
# from argparse import Namespace
from types import SimpleNamespace

# Include secondary dependencies here, since pyinstaller will miss them and
# the binary will fail at run-time.
import pkgutil 

HOST = config.settings['host']
MASTER_KEY = config.settings['master_key']
DATABASE_ID = config.settings['database_id']
CONTAINER_ID = config.settings['container_id']

logging.basicConfig(filename="cosmos-sync.log", level=logging.DEBUG)

logger = logging.getLogger('post-receive-bin')
#formatter = logging.Formatter('%(asctime)s - %(name)s - %(levelname)s - %(message)s')
#logger.setFormatter(formatter)

handler = RotatingFileHandler("cosmos-sync.log", maxBytes=20000, backupCount=5)
logger.addHandler(handler)
logger.propagate = False

def upsert_item(container, tguid, offset, size):
    logger.info('Upserting tguid = {0}'.format(tguid))

    logger.info('tguid: {0}'.format(tguid))
    logger.info('offset: {0}'.format(offset))
    logger.info('size: {0}'.format(size))

    # latest_offset = 0
    # with open('/tmp/{0}.txt'.format(tguid)) as f:
    #     latest_offset = int(f.readline())
    #     f.close()

    # logger.info('upsert_item: latest_offset = {0}, our offset = {1}'.format(latest_offset, offset))
    # if (offset < latest_offset):
    #     # This update is stale so skip it.  This is due to threading as order of upsert_item
    #     # calls at this point is no longer guaranteed.
    #     logger.warning("Our information is stale - skipping (latest_offset = {0}, this offset = {1})".format(latest_offset, offset))
    #     return

    try:
        logger.info('Checking to see if tguid exists...')
        read_item = container.read_item(item=tguid, partition_key='UploadStatus')
        logger.info('Found tguid')
        if (read_item['offset'] >= offset):
            logger.warning('Out of order call, continuing...')
            return
        logger.info('Updating found tguid with new offset')
        read_item['offset'] = offset
    except exceptions.CosmosHttpResponseError:
        logger.info('tguid not found')
        read_item = {
            'id' : tguid,
            'tguid' : tguid,
            'partitionKey' : 'UploadStatus',
            'offset' : offset,
            'size' : size
        }
    logger.info('Calling upsert_item')
    response = container.upsert_item(body=read_item)
    logger.info('Done calling upsert_item')
    
    logger.info('Upserted at {0}, new offset={1}'.format(datetime.datetime.now(), response['offset']))
    logger.info('Upsert success for tguid {0}!!'.format(tguid))

def init_db():
    client = cosmos_client.CosmosClient(HOST, {'masterKey': MASTER_KEY} )
    try:
        # setup database
        db = client.create_database_if_not_exists(id=DATABASE_ID)
        # setup container
        return db.create_container_if_not_exists(id=CONTAINER_ID, partition_key=PartitionKey(path='/partitionKey', kind='Hash'))

    except exceptions.CosmosHttpResponseError as e:
        logger.exception(e)

def post_receive(tguid, offset, size):
    try:
        logger.info('{0}, offset = {1}'.format(datetime.datetime.now(), offset))
        container = init_db()
        upsert_item(container, tguid, offset, size)
    except Exception as e:
        logger.exception(e)
        sys.exit(1)

try:
    logger.info('cosmos_sync with pid {0} now running...'.format(os.getpid()))
    lock_filename = 'cosmos-sync-bin.lock'
    if os.path.exists(lock_filename): os.remove(lock_filename)

    port = "5556"
    context = zmq.Context()
    socket = context.socket(zmq.PAIR)
    socket.bind("tcp://*:%s" % port)
    while True:
        logger.debug("Waiting for socket.recv()...")
        json_data = socket.recv()
        logger.debug("socket.recv(): received data: {0}".format(json_data))
        logger.info('Sending ack')
        socket.send_string('ACK')
        update = json.loads(json_data, object_hook=lambda d: SimpleNamespace(**d))
        logger.debug('tguid = {0}, offset = {1}, size = {2}'.format(update.tguid, update.offset, update.size))
        post_receive(update.tguid, update.offset, update.size)
        time.sleep(1e-1) # needed?
except Exception as e:
    logger.exception(e)
    sys.exit(1)
