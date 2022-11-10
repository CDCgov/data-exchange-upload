from tusclient import client

my_client = client.TusClient('https://as-bulk-upload-tusd.azurewebsites.net/files')

# create the uploader
uploader = my_client.uploader('10GB-test-file', metadata={'filename':'10GB-test-file'})

# upload the entire file chunk by chunk
uploader.upload()
