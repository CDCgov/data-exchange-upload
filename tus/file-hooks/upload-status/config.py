import os

queue_settings = {
    'storage_account_name': os.environ.get('AZURE_STORAGE_ACCOUNT'),
    'storage_account_key': os.environ.get('AZURE_STORAGE_KEY'),
    'queue_name': os.environ.get('QUEUE_NAME', 'cosmos-sink-queue')
}