import os

queue_settings = {
    'service_bus_connection_str': os.environ.get('SERVICE_BUS_CONNECTION_STR'),
    'queue_name': os.environ.get('QUEUE_NAME', 'processing-status-cosmos-db-queue')
}