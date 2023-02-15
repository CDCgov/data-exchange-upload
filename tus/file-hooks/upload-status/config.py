import os

settings = {
    'host': os.environ.get('ACCOUNT_HOST', 'https://tusd-state-db.documents.azure.com:443/'),
    'master_key': os.environ.get('ACCOUNT_KEY', 'yqQK3aEkquNYZ41oNKV6IGs50eb1kLhzN8OTRh4KaGGMRB5VHjZ4lnjRvAQI0EUIXB3Jm60ywRtHACDbDbMixA=='),
    'database_id': os.environ.get('COSMOS_DATABASE', 'UploadStatus'),
    'container_id': os.environ.get('COSMOS_CONTAINER', 'Items'),
}