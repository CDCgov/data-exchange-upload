import getopt
import sys, os
import subprocess
import psutil

import datetime

import zmq
import sys

import logging
from logging.handlers import RotatingFileHandler

logging.basicConfig(filename="cosmos-sync.log", level=logging.DEBUG)

logger = logging.getLogger('post-receive-bin')
handler = RotatingFileHandler("post-receive.log", maxBytes=20000, backupCount=5)
logger.addHandler(handler)
logger.propagate = False

def log(msg):
    print('[post-receive-bin] {0} {1}'.format(datetime.datetime.now(), msg))

# !!!!!!!!!!!!!!!!
# !!!!!!!!!!!!!!!!
# !!!!!!!!!!!!!!!!
# !!!!!!!!!!!!!!!!
# !!!!!!!!!!!!!!!!
# looks like when this method is called concurrently due to rapid updates than the socket connection
# is blocked for the second call.  would pub/sub or something other than pair type work better?  or,
# do we need some way to wait for previous call to complete before attempting to open the socket again?
# since with pair you can only have 1:1 client:server. 
# !!!!!!!!!!!!!!!!
# !!!!!!!!!!!!!!!!
# !!!!!!!!!!!!!!!!
# !!!!!!!!!!!!!!!!
# !!!!!!!!!!!!!!!!
# !!!!!!!!!!!!!!!!
def send_message_to_cosmos_sync(json_update):
    try:
        logger.debug('Entering send_message_to_cosmos_sync')
        port = "5556"
        context = zmq.Context()
        socket = context.socket(zmq.PAIR)
        logger.debug('send_message_to_cosmos_sync: calling socket.connect')
        socket.connect("tcp://localhost:%s" % port)
        #socket.setsockopt(zmq.LINGER, 100) # added
        logger.debug('send_message_to_cosmos_sync: calling socket.send_json')
        socket.send_json(json_update)
        log('INFO: Called socket.send_json(), waiting for ACK...')
        logger.debug('Called socket.send_json(), waiting for ACK...')
        socket.recv()
        log('INFO: ACK received, closing socket')
        logger.debug('ACK received, closing socket')
        socket.close() # added
    except Exception as e:
        log(str(e))
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
    log('INFO: cosmos_pids = {0}'.format(cosmos_pids))

    for pid in cosmos_pids:
        try:
            proc = psutil.Process(pid)
            if proc.status() == psutil.STATUS_ZOMBIE:
                log('WARNING: pid {0} is a Zombie process!'.format(pid))
            else:
                log('INFO: pid {0} appears to be running'.format(pid))
                return True
        except psutil.NoSuchProcess:
            log('ERROR: pid {0} not found'.format(pid))
    return False

def upsert_item(tguid, offset, size):
    log('INFO: Upserting tguid = {0}'.format(tguid))

    log('INFO: tguid: {0}'.format(tguid))
    log('INFO: offset: {0}'.format(offset))
    log('INFO: size: {0}'.format(size))

    log('INFO: Sending update to cosmos-sync-bin...')
    update = {
        'tguid': tguid,
        'offset': offset,
        'size': size
    }
    send_message_to_cosmos_sync(update)

def post_receive(tguid, offset, size):
    try:
        log('INFO: python version = {0}'.format(sys.version))

        processname = 'cosmos-sync-bin'
        cosmos_sync_running = is_cosmos_sync_running(processname)
        if cosmos_sync_running == True:
            log('INFO: {0} is running'.format(processname))
        else:
            start_in_progress_filename = './{0}.lock'.format(processname)
            start_in_progress = os.path.exists(start_in_progress_filename)
            if start_in_progress:
                log('INFO: {0} in process of starting...'.format(processname))
            else:
                log('INFO: Starting {0}...'.format(processname))
                # create empty lock file
                with open(start_in_progress_filename, 'w'):
                    pass
                # Spawn cosmos-sync-bin as a new process and don't wait for it
                proc = subprocess.Popen('./' + processname, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
                log('INFO: Started {0} with pid = {1}'.format(processname, proc.pid))

        log('INFO: post_receive_bin: {0}, offset = {1}'.format(datetime.datetime.now(), offset))
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