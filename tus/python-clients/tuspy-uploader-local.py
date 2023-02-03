from tusclient import client
from tusclient.exceptions import TusCommunicationError

my_client = client.TusClient('http://0.0.0.0:1080/files/')

# create the uploader
uploader = my_client.uploader('flower.jpeg', metadata={'filename':'flower.jpeg','meta_destination_id':'ndlp','meta_ext_event':'ri','meta_ext_source':'IZGW'})

# upload the entire file
try:
    uploader.upload()
except TusCommunicationError as error:
    print('TusCommunicationError: ' + str(error.response_content.decode('UTF-8')))
