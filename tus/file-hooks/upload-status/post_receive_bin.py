import getopt
import sys
import azure.cosmos.cosmos_client as cosmos_client
import azure.cosmos.exceptions as exceptions
from azure.cosmos.partition_key import PartitionKey

import config
import datetime

# Include secondary dependencies here, since pyinstaller will miss them and
# the binary will fail at run-time.
import pkgutil 

HOST = config.settings['host']
MASTER_KEY = config.settings['master_key']
DATABASE_ID = config.settings['database_id']
CONTAINER_ID = config.settings['container_id']

def upsert_item(container, tguid, offset, size):
    print('INFO: Upserting tguid = {0}'.format(tguid))

    print('INFO: tguid: {0}'.format(tguid))
    print('INFO: offset: {0}'.format(offset))
    print('INFO: size: {0}'.format(size))

    latest_offset = 0
    with open('/tmp/{0}.txt'.format(tguid)) as f:
        latest_offset = int(f.readline())
        f.close()

    print('INFO: upsert_item: latest_offset = {0}, our offset = {1}'.format(latest_offset, offset))
    if (offset < latest_offset):
        # This update is stale so skip it.  This is due to threading as order of upsert_item
        # calls at this point is no longer guaranteed.
        print("Our information is stale - skipping (latest_offset = {0}, this offset = {1})".format(latest_offset, offset))
        return

    try:
        print('INFO: Checking to see if tguid exists...')
        read_item = container.read_item(item=tguid, partition_key='UploadStatus')
        print('INFO: Found tguid')
        if (read_item['offset'] >= offset):
            print('WARNING: Out of order call, continuing...')
            return
        print('INFO: Updating found tguid with new offset')
        read_item['offset'] = offset
    except exceptions.CosmosHttpResponseError:
        print('INFO: tguid not found')
        read_item = {
            'id' : tguid,
            'tguid' : tguid,
            'partitionKey' : 'UploadStatus',
            'offset' : offset,
            'size' : size
        }
    print('INFO: Calling upsert_item')
    response = container.upsert_item(body=read_item)
    print('INFO: Done calling upsert_item')
    
    print('INFO: Upserted at {0}, new offset={1}'.format(datetime.datetime.now(), response['offset']))
    print('INFO: Upsert success!!')

def init_db():
    client = cosmos_client.CosmosClient(HOST, {'masterKey': MASTER_KEY} )
    try:
        # setup database for this sample
        db = client.create_database_if_not_exists(id=DATABASE_ID)
        # setup container for this sample
        return db.create_container_if_not_exists(id=CONTAINER_ID, partition_key=PartitionKey(path='/partitionKey', kind='Hash'))

    except exceptions.CosmosHttpResponseError as e:
        print('Error: {0}'.format(e.message))

def post_receive(tguid, offset, size):
    try:
        print('INFO: post_receive_bin: {0}, offset = {1}'.format(datetime.datetime.now(), offset))
        container = init_db()
        upsert_item(container, tguid, offset, size)
    except Exception as e:
        print(e)
        sys.exit(1)

def main(argv):
    tus_id = ''
    offset = ''
    size = ''
    opts, args = getopt.getopt(argv,"hios:",["id=","offset=","size="])
    for opt, arg in opts:
        if opt == '-h':
            print ('post_receive_bin.py -i id -o offset -s size')
            sys.exit()
        elif opt in ("-i", "--id"):
            tus_id = arg
        elif opt in ("-o", "--offset"):
            offset = arg
        elif opt in ("-s", "--size"):
            size = arg
    try:
        post_receive(tus_id, int(offset), int(size))
    except Exception as e:
        print(e)
        sys.exit(1)

if __name__ == "__main__":
    main(sys.argv[1:])