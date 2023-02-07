import tus

with open('flower.jpeg', 'rb') as f:
    tus.upload(
    	f,
        'http://0.0.0.0:1080/files/',
        metadata={'filename':'flower.jpeg'})
