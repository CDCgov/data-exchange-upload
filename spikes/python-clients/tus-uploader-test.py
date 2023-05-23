import tus

with open('flower.jpeg', 'rb') as f:
    tus.upload(
    	f,
        'https://as-bulk-upload-tusd.azurewebsites.net/files/',
        metadata={'filename':'flower.jpeg'},
        chunk_size=2000)