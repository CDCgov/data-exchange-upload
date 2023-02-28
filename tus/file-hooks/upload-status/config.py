import os

settings = {
    'host': os.environ.get('ACCOUNT_HOST', 'https://tusd-state-db.documents.azure.com:443/'),
    'master_key': os.environ.get('ACCOUNT_KEY', 'yqQK3aEkquNYZ41oNKV6IGs50eb1kLhzN8OTRh4KaGGMRB5VHjZ4lnjRvAQI0EUIXB3Jm60ywRtHACDbDbMixA=='),
    'database_id': os.environ.get('COSMOS_DATABASE', 'UploadStatus'),
    'container_id': os.environ.get('COSMOS_CONTAINER', 'Items'),
}

queue_settings = {
    'storage_connection_string': os.environ.get('AZURE_STORAGE_CONNECTION_STRING', 'DefaultEndpointsProtocol=https;AccountName=dataexchangedev;AccountKey=lVvJbZ5J+SvLvWpUMwybFKnqYs57J4EF+HBvWTUo9GAHsLheFRWHOxXmVmy2Ojy7m/W8qBbgXIoe+AStzh0IdQ==;EndpointSuffix=core.windows.net'),
    'queue_name': os.environ.get('QUEUE_NAME', 'cosmos-sink-queue')
}