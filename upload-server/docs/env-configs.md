# Environment Configuration for Upload Server

This document outlines the environment variables that can be configured for the Upload Server application. The `.env.example` file contains the default values where applicable.

## How to Set Environment Variables

1. Copy the `.env.example` file to `.env` in the same directory.
2. Modify the variables as needed.
3. Ensure that all required variables are set.

## Common Configs

### Logging and Environment

| Variable Name               | Required | Default Value | Description                                                 |
|-----------------------------|----------|---------------|------------------------------------------------------------|
| `LOGGER_DEBUG_ON`           | No       | None          | Enable or disable debug logging. Values: `INFO` or `DEBUG` |
| `ENVIRONMENT`               | No       | `development` | The environment mode (`development`, `production`, etc.).  |

### Server Configs

| Variable Name                  | Required | Default Value       | Description                                                    |
|--------------------------------|----------|---------------------|----------------------------------------------------------------|
| `SERVER_PROTOCOL`              | No       | `http`              | The protocol used by the server (`http`, `https`).             |
| `SERVER_HOSTNAME`              | No       | `localhost`         | The hostname of the server.                                    |
| `SERVER_PORT`                  | No       | `8080`              | The port on which the server runs.                             |
| `TUSD_HANDLER_BASE_PATH`       | No       | `/files`            | The base path for handling tusd requests.                      |
| `TUSD_HANDLER_INFO_PATH`       | No       | `/files/info`       | Path for tusd info handler.                                    |
| `UPLOAD_CONFIG_PATH`           | No       | `../upload-configs` | Path to the upload configuration files.                        |
| `EVENT_MAX_RETRY_COUNT`        | No       | `3`                 | Maximum number of retry attempts for event processing.         |
| `METRICS_LABELS_FROM_MANIFEST` | No       | `false`             | Enable or disable metrics labels generation from manifest.     |
| `TUS_UPLOAD_PREFIX`            | No       | `/uploads`          | Prefix path for tus uploads.                                   |

### User Interface Configs

| Variable Name              | Required | Default Value   | Description                                                    |
|----------------------------|----------|-----------------|----------------------------------------------------------------|
| `UI_PORT`                  | No       | `8081`          | The port for the local UI.                                     |
| `CSRF_TOKEN`               | Yes      | None            | CSRF token used for authentication. 32 byte string.            |

### Redis Configs

| Variable Name              | Required | Default Value   | Description                                                   |
|----------------------------|----------|-----------------|---------------------------------------------------------------|
| `REDIS_CONNECTION_STRING`  | No       | None            | Connection string for Redis.                                  |

### OAuth Configs

| Variable Name               | Required | Default Value   | Description                                                    |
|-----------------------------|----------|-----------------|----------------------------------------------------------------|
| `OAUTH_AUTH_ENABLED`        | No       | `false`         | Enable OAuth authentication.                                   |
| `OAUTH_INTROSPECTION_URL`   | No       | None            | OAuth introspection URL.                                       |
| `OAUTH_ISSUER_URL`          | No       | None            | OAuth issuer URL.                                              |
| `OAUTH_REQUIRED_SCOPES`     | No       | None            | Required scopes for OAuth.                                     |

## Upload Location Configs

### Local File System Configs
| Variable Name               | Required | Default Value   | Description                                                   |
|-----------------------------|----------|-----------------|---------------------------------------------------------------|
| `LOCAL_FOLDER_UPLOADS_TUS`  | No       | `./uploads`     | Local file system upload directory.                           |

### Azure Storage Configs

| Variable Name                       | Required | Default Value | Description                                                    |
|------------------------------------ |----------|---------------|----------------------------------------------------------------|
| `AZURE_STORAGE_ACCOUNT`             | No       | None          | Azure storage account name.                                    |
| `AZURE_STORAGE_KEY`                 | No       | None          | Azure storage account key.                                     |
| `AZURE_ENDPOINT`                    | No       | None          | Azure storage endpoint URL.                                    |
| `AZURE_TENANT_ID`                   | No       | None          | Azure Active Directory tenant ID.                              |
| `AZURE_CLIENT_ID`                   | No       | None          | Azure Active Directory client ID.                              |
| `AZURE_CLIENT_SECRET`               | No       | None          | Azure Active Directory client secret.                          |
| `TUS_AZURE_CONTAINER_NAME`          | No       | None          | Container name for tus Azure integration.                      |
| `DEX_MANIFEST_CONFIG_CONTAINER_NAME`| No       | None          | Container name for manifest configurations.                    |

## S3 Storage Configs

| Variable Name                          | Required | Default Value | Description                                                                                  |
|---------------------------------------- |----------|---------------|----------------------------------------------------------------------------------------------|
| `S3_ENDPOINT`                           | Yes      | None          | S3-compatible storage endpoint URL.                                                           |
| `S3_BUCKET_NAME`                        | Yes      | None          | Name of the S3 bucket for storage.                                                            |
| `AWS_REGION`                            | Yes      | None          | AWS region for the S3 bucket.                                                                 |
| `AWS_ACCESS_KEY_ID`                     | Yes      | None          | AWS access key ID for authenticating access to the bucket.                                    |
| `AWS_SECRET_ACCESS_KEY`                 | Yes      | None          | AWS secret access key for authenticating access to the bucket.                                |
| `DEX_MANIFEST_CONFIG_BUCKET_NAME` *     | No       | None          | Bucket name for storing manifest configurations, if different from the primary bucket.        |
| `DEX_S3_MANIFEST_CONFIG_FOLDER_NAME` *  | No       | None          | Folder name within the primary bucket for storing manifest configurations.                    |

> `*` **Note:** Only set one of `DEX_MANIFEST_CONFIG_BUCKET_NAME` or `DEX_S3_MANIFEST_CONFIG_FOLDER_NAME`. Use `DEX_MANIFEST_CONFIG_BUCKET_NAME` if the manifest configurations are in a different bucket; otherwise, use `DEX_S3_MANIFEST_CONFIG_FOLDER_NAME` if they are stored in a different folder within the same bucket.

# Report Location Configs

### Local File System Report Directory

| Variable Name                 | Required | Default Value      | Description                                      |
|-------------------------------|----------|--------------------|--------------------------------------------------|
| `LOCAL_REPORTS_FOLDER`        | No       | `./upload/reports` | Local folder path where reports will be stored.  |

### Azure Report Queue

| Variable Name                  | Required | Default Value   | Description                                                   |
|--------------------------------|----------|-----------------|---------------------------------------------------------------|
| `REPORTER_CONNECTION_STRING`   | No       | None            | Connection string for Azure queue or service bus.             |
| `REPORTER_QUEUE`               | No       | None            | Queue name for sending reports to the Azure queue system.     |
| `REPORTER_TOPIC`               | No       | None            | Topic name for the Azure queue or service bus.                |




## Event Publish/Subscribe Location Configs

## Local File System Event Directory

| Variable Name          | Required | Default Value       | Description                                                |
|------------------------|----------|---------------------|------------------------------------------------------------|
| `LOCAL_EVENTS_FOLDER`  | No       | `./uploads/events`  | Local folder path where events will be stored.             |


### Azure Event Publisher Queue

| Variable Name                 | Required | Default Value   | Description                                                   |
|-------------------------------|----------|-----------------|---------------------------------------------------------------|
| `PUBLISHER_CONNECTION_STRING` | No       | None            | Connection string for the Azure event publisher queue.        |
| `PUBLISHER_TOPIC`             | No       | None            | Topic name for publishing events to the Azure queue system.   |

### Azure Event Subscriber Queue

| Variable Name                   | Required | Default Value   | Description                                                   |
|---------------------------------|----------|-----------------|---------------------------------------------------------------|
| `SUBSCRIBER_CONNECTION_STRING`  | No       | None            | Connection string for the Azure event subscriber queue.       |
| `SUBSCRIBER_TOPIC`              | No       | None            | Topic name for subscribing to events in the Azure queue.      |
| `SUBSCRIBER_SUBSCRIPTION`       | No       | None            | Subscription name for the Azure event subscriber.             |

## File Delivery Target Configs

### EDAV Delivery Target

#### Local File System EDAV Directory

| Variable Name          | Required | Default Value       | Description                                                |
|------------------------|----------|---------------------|------------------------------------------------------------|
| `LOCAL_EDAV_FOLDER`    | No       | `./uploads/edav`    | Local folder path for EDAV file deliveries.                |

#### Azure EDAV Container

| Variable Name                   | Required | Default Value           | Description                                                    |
|---------------------------------|----------|-------------------------|----------------------------------------------------------------|
| `EDAV_STORAGE_ACCOUNT`          | No       | None                    | Azure storage account name for EDAV deliveries.                |
| `EDAV_STORAGE_KEY`              | No       | None                    | Azure storage account key for EDAV deliveries.                 |
| `EDAV_TENANT_ID`                | No       | None                    | Azure Active Directory tenant ID for EDAV deliveries.          |
| `EDAV_CLIENT_ID`                | No       | None                    | Azure Active Directory client ID for EDAV deliveries.          |
| `EDAV_CLIENT_SECRET`            | No       | None                    | Azure Active Directory client secret for EDAV deliveries.      |
| `EDAV_ENDPOINT`                 | No       | None                    | Azure endpoint for EDAV storage.                               |
| `EDAV_CHECKPOINT_CONTAINER_NAME`| No       | `edav-checkpoint`       | Azure container name for checkpoint data related to EDAV.      |

#### S3 EDAV Bucket

| Variable Name                 | Required | Default Value   | Description                                                   |
|-------------------------------|----------|-----------------|---------------------------------------------------------------|
| `EDAV_S3_ENDPOINT`            | Yes      | None            | S3-compatible storage endpoint for EDAV deliveries.            |
| `EDAV_S3_BUCKET_NAME`         | Yes      | None            | Name of the S3 bucket for EDAV deliveries.                     |

### EHDI Delivery Target

#### Local File System EHDI Directory

| Variable Name          | Required | Default Value       | Description                                                |
|------------------------|----------|---------------------|------------------------------------------------------------|
| `LOCAL_EHDI_FOLDER`    | No       | `./uploads/ehdi`    | Local folder path for EHDI file deliveries.                |


#### Azure EHDI Container

| Variable Name                   | Required | Default Value           | Description                                                    |
|---------------------------------|----------|-------------------------|----------------------------------------------------------------|
| `EHDI_STORAGE_ACCOUNT`          | No       | None                    | Azure storage account name for EHDI deliveries.                |
| `EHDI_STORAGE_KEY`              | No       | None                    | Azure storage account key for EHDI deliveries.                 |
| `EHDI_TENANT_ID`                | No       | None                    | Azure Active Directory tenant ID for EHDI deliveries.          |
| `EHDI_CLIENT_ID`                | No       | None                    | Azure Active Directory client ID for EHDI deliveries.          |
| `EHDI_CLIENT_SECRET`            | No       | None                    | Azure Active Directory client secret for EHDI deliveries.      |
| `EHDI_ENDPOINT`                 | No       | None                    | Azure endpoint for EHDI storage.                               |
| `EHDI_CHECKPOINT_CONTAINER_NAME`| No       | `ehdi-checkpoint`       | Azure container name for checkpoint data related to EHDI.      |


#### S3 EHDI Bucket

| Variable Name                 | Required | Default Value   | Description                                                   |
|-------------------------------|----------|-----------------|---------------------------------------------------------------|
| `EHDI_S3_ENDPOINT`            | No       | None            | S3-compatible storage endpoint for EHDI deliveries.           |
| `EHDI_S3_BUCKET_NAME`         | No       | None            | Name of the S3 bucket for EHDI deliveries.                    |

### EICR Delivery Target

#### Local File System EICR Directory

| Variable Name          | Required | Default Value       | Description                                                |
|------------------------|----------|---------------------|------------------------------------------------------------|
| `LOCAL_EICR_FOLDER`    | No       | `./uploads/eicr`    | Local folder path for EICR file deliveries.                |


#### Azure EICR Container

| Variable Name                    | Required | Default Value           | Description                                                    |
|----------------------------------|----------|-------------------------|----------------------------------------------------------------|
| `EICR_STORAGE_ACCOUNT`           | No       | None                    | Azure storage account name for EICR deliveries.                |
| `EICR_STORAGE_KEY`               | No       | None                    | Azure storage account key for EICR deliveries.                 |
| `EICR_TENANT_ID`                 | No       | None                    | Azure Active Directory tenant ID for EICR deliveries.          |
| `EICR_CLIENT_ID`                 | No       | None                    | Azure Active Directory client ID for EICR deliveries.          |
| `EICR_CLIENT_SECRET`             | No       | None                    | Azure Active Directory client secret for EICR deliveries.      |
| `EICR_ENDPOINT`                  | No       | None                    | Azure endpoint for EICR storage.                               |
| `EICR_CHECKPOINT_CONTAINER_NAME` | No       | `eicr-checkpoint`       | Azure container name for checkpoint data related to EICR.      |


#### S3 EICR Bucket

| Variable Name                 | Required | Default Value   | Description                                                   |
|-------------------------------|----------|-----------------|---------------------------------------------------------------|
| `EICR_S3_ENDPOINT`            | No       | None            | S3-compatible storage endpoint for EICR deliveries.           |
| `EICR_S3_BUCKET_NAME`         | No       | None            | Name of the S3 bucket for EICR deliveries.                    |

### NCIRD Delivery Target

#### Local File System NCIRD Directory

| Variable Name          | Required | Default Value       | Description                                                |
|------------------------|----------|---------------------|------------------------------------------------------------|
| `LOCAL_NCIRD_FOLDER`   | No       | `./uploads/ncird`   | Local folder path for NCIRD file deliveries.               |

#### Azure NCIRD Container

| Variable Name                     | Required | Default Value           | Description                                                    |
|-----------------------------------|----------|-------------------------|----------------------------------------------------------------|
| `NCIRD_STORAGE_ACCOUNT`           | No       | None                    | Azure storage account name for NCIRD deliveries.               |
| `NCIRD_STORAGE_KEY`               | No       | None                    | Azure storage account key for NCIRD deliveries.                |
| `NCIRD_TENANT_ID`                 | No       | None                    | Azure Active Directory tenant ID for NCIRD deliveries.         |
| `NCIRD_CLIENT_ID`                 | No       | None                    | Azure Active Directory client ID for NCIRD deliveries.         |
| `NCIRD_CLIENT_SECRET`             | No       | None                    | Azure Active Directory client secret for NCIRD deliveries.     |
| `NCIRD_ENDPOINT`                  | No       | None                    | Azure endpoint for NCIRD storage.                              |
| `NCIRD_CHECKPOINT_CONTAINER_NAME` | No       | `ncird-checkpoint`      | Azure container name for checkpoint data related to NCIRD.     |

#### S3 NCIRD Bucket

| Variable Name                | Required | Default Value   | Description                                                   |
|------------------------------|----------|-----------------|---------------------------------------------------------------|
| `NCIRD_S3_ENDPOINT`          | No       | None            | S3-compatible storage endpoint for NCIRD deliveries.          |
| `NCIRD_S3_BUCKET_NAME`       | No       | None            | Name of the S3 bucket for NCIRD deliveries.                   |

