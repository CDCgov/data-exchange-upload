from tusclient import client
from tusclient.exceptions import TusCommunicationError
import uuid

my_client = client.TusClient('https://as-bulk-upload-tusd.azurewebsites.net/files')

# create the uploader
uploader = my_client.uploader('flower.jpeg',
    metadata={
        'filename':'flower.jpeg',
        'meta_destination_id':'ndlp',
        'meta_ext_event':'ri',
        'meta_ext_source':'IZGW1',
        'meta_ext_entity':'DY2',
        'meta_username':'ygj6@cdc.gov',
        'meta_ext_objectkey':str(uuid.uuid4()),
        'meta_ext_filename':'flower.jpeg'
        }
    )

# upload the entire file
try:
    uploader.upload()
except TusCommunicationError as error:
    print('TusCommunicationError: ' + str(error.response_content.decode('UTF-8')))