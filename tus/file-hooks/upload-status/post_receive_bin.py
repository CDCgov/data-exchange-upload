import getopt
import sys, os
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
    print('\nUpserting tguid = {0}\n'.format(tguid))

    print('tguid: {0}'.format(tguid))
    print('offset: {0}'.format(offset))
    print('size: {0}'.format(size))

    latest_offset = 0
    with open('/tmp/{0}.txt'.format(tguid)) as f:
        latest_offset = int(f.readline())
        f.close()

    print('upsert_item: latest_offset = {0}, our offset = {1}'.format(latest_offset, offset))
    if (offset < latest_offset):
        # This update is stale so skip it.  This is due to threading as order of upsert_item
        # calls at this point is no longer guaranteed.
        print("Our information is stale - skipping (latest_offset = {0}, this offset = {1})".format(latest_offset, offset))
        return

    try:
        print('Checking to see if tguid exists...')
        read_item = container.read_item(item=tguid, partition_key='UploadStatus')
        print('Found tguid')
        if (read_item['offset'] >= offset):
            print('Out of order call, continuing...')
            return
        print('Updating found tguid with new offset')
        read_item['offset'] = offset
    except exceptions.CosmosHttpResponseError:
        print('tguid not found')
        read_item = {
            'id' : tguid,
            'tguid' : tguid,
            'partitionKey' : 'UploadStatus',
            'offset' : offset,
            'size' : size
        }
    print('Calling upsert_item')
    response = container.upsert_item(body=read_item)
    print('Done calling upsert_item')
    
    print('Upserted at {0}, new offset={1}'.format(datetime.datetime.now(), response['offset']))
    print('Upsert success!!')

def init_db():
    client = cosmos_client.CosmosClient(HOST, {'masterKey': MASTER_KEY} )
    try:
        # setup database for this sample
        db = client.create_database_if_not_exists(id=DATABASE_ID)
        # setup container for this sample
        container = db.create_container_if_not_exists(id=CONTAINER_ID, partition_key=PartitionKey(path='/partitionKey', kind='Hash'))

        # setup container
        # try:
        #     container = db.create_container(id=CONTAINER_ID, partition_key=PartitionKey(path='/partitionKey'))
        #     print('Container with id \'{0}\' created'.format(CONTAINER_ID))

        # except exceptions.CosmosResourceExistsError:
        #     container = db.get_container_client(CONTAINER_ID)
        #     print('Container with id \'{0}\' was found'.format(CONTAINER_ID))

        # scale_container(container)
        
        return container

    except exceptions.CosmosHttpResponseError as e:
        print('\nrun_sample has caught an error. {0}'.format(e.message))

#def main():
def post_receive(tguid, offset, size):
    try:
        env_tguid = os.getenv('TUS_ID')
        env_offset = int(os.getenv('TUS_OFFSET'))
        env_size = int(os.getenv('TUS_SIZE'))

        # print('*** Compare env_tguid = {0}, tguid from bash = {1}'.format(env_tguid, tguid))
        print('*** Compare env_offset = {0}, offset from bash = {1}'.format(env_offset, offset))
        # print('*** Compare env_size = {0}, size from bash = {1}'.format(env_size, size))

        print('post_receive_bin: {0}, offset = {1}'.format(datetime.datetime.now(), offset))
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