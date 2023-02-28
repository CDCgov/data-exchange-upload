import getopt
import sys, os
import subprocess
import psutil
import config

import datetime

from azure.storage.queue import QueueServiceClient
from azure.core.exceptions import ResourceExistsError

import json

import logging

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

connect_str = config.queue_settings['storage_connection_string']
q_name = config.queue_settings['queue_name']

def get_queue_client():
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
        get_queue_client().send_message(json_update)
    except Exception as e:
        logger.exception(e)

def is_cosmos_sync_running(cosmos_sync_proc_name):
    ps_result = os.popen("ps -ef").read()
    lines = ps_result.splitlines()
    
    cosmos_items = [s for s in lines if cosmos_sync_proc_name in s]
    cosmos_pids = []
    for item in cosmos_items:
        columns = item.split()
        if len(columns) == 4:
            cosmos_pids.append(int(columns[0]))
    logger.info('cosmos_pids = {0}'.format(cosmos_pids))

    for pid in cosmos_pids:
        try:
            proc = psutil.Process(pid)
            if proc.status() == psutil.STATUS_ZOMBIE:
                logger.warning('pid {0} is a Zombie process!'.format(pid))
            else:
                logger.info('pid {0} appears to be running'.format(pid))
                return True
        except psutil.NoSuchProcess:
            logger.error('pid {0} not found'.format(pid))
    return False

def upsert_item(tguid, offset, size):
    logger.info('Upserting tguid = {0}'.format(tguid))

    logger.info('tguid: {0}'.format(tguid))
    logger.info('offset: {0}'.format(offset))
    logger.info('size: {0}'.format(size))

    logger.info('Sending update to cosmos-sync-bin...')
    update = {
        'tguid': tguid,
        'offset': offset,
        'size': size
    }
    json_update = json.dumps(update)
    send_message_to_cosmos_sync(json_update)

def post_receive(tguid, offset, size):
    try:
        logger.info('python version = {0}'.format(sys.version))

        processname = 'cosmos-sync-bin'
        cosmos_sync_running = is_cosmos_sync_running(processname)
        if cosmos_sync_running == True:
            logger.info('{0} is running'.format(processname))
        else:
            start_in_progress_filename = './{0}.lock'.format(processname)
            start_in_progress = os.path.exists(start_in_progress_filename)
            if start_in_progress:
                logger.info('{0} in process of starting...'.format(processname))
            else:
                logger.info('Starting {0}...'.format(processname))
                # create empty lock file
                with open(start_in_progress_filename, 'w'):
                    pass
                # Spawn cosmos-sync-bin as a new process and don't wait for it
                proc = subprocess.Popen('./' + processname, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
                logger.info('Started {0} with pid = {1}'.format(processname, proc.pid))

        logger.info('post_receive_bin: {0}, offset = {1}'.format(datetime.datetime.now(), offset))
        upsert_item(tguid, offset, size)
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
            print('post_receive_bin.py -i id -o offset -s size')
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