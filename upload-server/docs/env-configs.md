# Environment Configuration for Upload Server

This document outlines the environment variables that can be configured for the Upload Server application. The `.env.example` file contains the default values where applicable.

## How to Set Environment Variables

1. Copy the `.env.example` file to `.env` in the same directory.
2. Modify the variables as needed.
3. Ensure that all required variables are set. In Common Configs, required variables are needed to start the servers. For other sections, required variables refers to using that features or connection.

## Common Configs

There are no required values to start the tus server or the tus client UI. However, it is highly recommended that you create a new `CSRF_TOKEN`, if you using the UI client for anything other than development.

### Logging and Environment

| Variable Name     | Required | Default Value | Description                                               |
|-------------------|----------|---------------|-----------------------------------------------------------|
| `LOGGER_DEBUG_ON` | No       | `INFO`        | Enable or disable DEBUG logging level                     |
| `ENVIRONMENT`     | No       | `DEV`         | Environment for the service (`DEV`, `TST`, `STG`, `PROD`) |

### Server Configs

| Variable Name                  | Required | Default Value                                | Description                                                                                |
|--------------------------------|----------|----------------------------------------------|--------------------------------------------------------------------------------------------|
| `SERVER_PROTOCOL`              | No       | `http`                                       | Protocol used by the server (`http`, `https`)                                              |
| `SERVER_HOSTNAME`              | No       | `localhost`                                  | Hostname of the server                                                                     |
| `SERVER_PORT`                  | No       | `8080`                                       | Port on which the server runs                                                              |
| `TUSD_HANDLER_BASE_PATH`       | No       | `/files/`                                    | URL path for handling tusd upload requests                                                 |
| `TUSD_HANDLER_INFO_PATH`       | No       | `/info/`                                     | URL Path for handling tusd info requests                                                   |
| `EVENT_MAX_RETRY_COUNT`        | No       | `3`                                          | Maximum number of retry attempts for event processing                                      |
| `METRICS_LABELS_FROM_MANIFEST` | No       | `data_stream_id,data_stream_route,sender_id` | String separated list of keys from the sender manifest config to count in the metrics      |
| `TUS_UPLOAD_PREFIX`            | No       | `tus-prefix`                                 | Relative file system path to the tus uploads directory within the storage backend location |

### User Interface Configs

| Variable Name  | Required | Default Value                                                      | Description                                         |
|----------------|----------|--------------------------------------------------------------------|-----------------------------------------------------|
| `UI_PORT`      | No       | `8081`                                                             | Port on which the UI client runs                    |
| `CSRF_TOKEN` * | No *     | `1qQBJumxRABFBLvaz5PSXBcXLE84viE42x4Aev359DvLSvzjbXSme3whhFkESatW` | CSRF token used for authentication (32 byte string) |

> `*` **Note:** The default is for development purposes only. You should replace this with a new string, you can generate a 32 byte string [here](https://generate-random.org/encryption-key-generator?count=1&bytes=32&cipher=aes-256-cbc&string=&password=).

## Redis Configs

Redis is used for handling distributed file locking across multiple upload servers. (see [Configuring Distributed File Locking with Redis](../README.md#configuring-distributed-file-locking-with-redis) for more information)

| Variable Name             | Required | Default Value | Description                             |
|---------------------------|----------|---------------|-----------------------------------------|
| `REDIS_CONNECTION_STRING` | Yes       | None          | Connection string to the Redis instance |

## OAuth Configs

OAuth token verification is used to secure the `/files/` and `/info/` UPLOAD API endpoints. (see [Configuring OAuth Token Verification Middleware](../README.md#configuring-oauth-token-verification-middleware))

| Variable Name             | Required | Default Value | Description                                          |
|---------------------------|----------|---------------|------------------------------------------------------|
| `OAUTH_AUTH_ENABLED`      | Yes      | `false`       | Enable or disable OAuth token verification           |
| `OAUTH_ISSUER_URL`        | Yes      | None          | URL of the OAuth token issuer                        |
| `OAUTH_REQUIRED_SCOPES`   | Yes      | None          | Space-separated list of required scopes              |
| `OAUTH_SESSION_KEY`       | Yes      | None          | Unique value to be used to hash a user session cookie.  Recommended to be at least 32 bytes long.  **Value is sensative and should not be checked into source control.**              |
| `OAUTH_SESSION_DOMAIN`    | No       | None          | Value used to set the Domain setting of the user session cookie.  Useful when the server and UI are on different subdomains.              |
| `OAUTH_INTROSPECTION_URL` | No       | None          | URL for OAuth introspection (used for opaque tokens) |

## Upload Location Configs

### Local File System Configs

| Variable Name              | Required | Default Value       | Description                                                              |
|----------------------------|----------|---------------------|--------------------------------------------------------------------------|
| `LOCAL_FOLDER_UPLOADS_TUS` | No       | `./uploads`         | Relative file system path to the base upload directory                   |
| `UPLOAD_CONFIG_PATH`       | No       | `../upload-configs` | Relative file system path to the sender manifest configuration directory |

### Azure Storage Configs

| Variable Name                        | Required | Default Value | Description                                            |
|--------------------------------------|----------|---------------|--------------------------------------------------------|
| `AZURE_STORAGE_ACCOUNT`              | Yes      | None          | Azure storage account name                             |
| `AZURE_STORAGE_KEY`                  | Yes      | None          | Azure storage account private access key or SAS token  |
| `AZURE_ENDPOINT`                     | Yes      | None          | Azure storage endpoint URL                             |
| `TUS_AZURE_CONTAINER_NAME`           | Yes      | None          | Container name for tus base upload storage             |
| `AZURE_TENANT_ID`                    | No       | None          | Service principal tenant ID                            |
| `AZURE_CLIENT_ID`                    | No       | None          | Service principal client ID                            |
| `AZURE_CLIENT_SECRET`                | No       | None          | Service principal client secret                        |
| `DEX_MANIFEST_CONFIG_CONTAINER_NAME` | No       | None          | Container name for sender manifest configuration files |

### S3 Storage Configs

| Variable Name                           | Required | Default Value | Description                                                                             |
|-----------------------------------------|----------|---------------|-----------------------------------------------------------------------------------------|
| `S3_ENDPOINT`                           | Yes      | None          | S3-compatible storage endpoint URL, must start with `http` or `https`                   |
| `S3_BUCKET_NAME`                        | Yes      | None          | Bucket name for  tus base upload storage                                                |
| `AWS_ACCESS_KEY_ID` *                   | Yes *    | None          | Username or user ID of the user or service account to access the bucket                 |
| `AWS_SECRET_ACCESS_KEY` *               | Yes *    | None          | Password or private key to the access bucket                                            |
| `AWS_SESSION_TOKEN` *                   | No       | None          | Session token for authentication (used for short lived keys)                            |
| `AWS_REGION` *                          | Yes *    | None          | Region of the S3 bucket                                                                 |
| `AWS_PROFILE` *                         | Yes *    | None          | Profile name of the AWS CLI profile to use                                              |
| `DEX_MANIFEST_CONFIG_BUCKET_NAME` **    | No       | None          | Bucket name for the sender manifest configurations, if different from the upload bucket |
| `DEX_S3_MANIFEST_CONFIG_FOLDER_NAME` ** | No       | None          | Directory name for the sender manifest configuration files within the upload bucket     |

> `*` **Note:** AWS Authentication can be handled using the standard environment variables `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`, and (optionally) `AWS_SESSION_TOKEN` or using profiles in AWS credential and configuration setting files with AWS CLI. If you are using profiles and the one you need is not `[default]`, you will need to set `AWS_PROFILE` to the correct profile name.
>
> `**` **Note:** Only set one of `DEX_MANIFEST_CONFIG_BUCKET_NAME` or `DEX_S3_MANIFEST_CONFIG_FOLDER_NAME`. Use `DEX_MANIFEST_CONFIG_BUCKET_NAME` if the manifest configurations are in a different bucket; otherwise, use `DEX_S3_MANIFEST_CONFIG_FOLDER_NAME` if they are stored in a different folder within the same bucket. If neither is set, the sender manifest configuration files on the local file system will be used.

## Report Location Configs

### Local File System Report Directory

| Variable Name                 | Required | Default Value      | Description                                      |
|-------------------------------|----------|--------------------|--------------------------------------------------|
| `LOCAL_REPORTS_FOLDER`        | No       | `./upload/reports` | Relative file system path to reports directory   |

### Azure Report Queue

| Variable Name                  | Required | Default Value   | Description                                                       |
|--------------------------------|----------|-----------------|-------------------------------------------------------------------|
| `REPORTER_CONNECTION_STRING`   | Yes      | None            | Azure connection string with credential to the queue or topic     |
| `REPORTER_QUEUE` *             | Yes *    | None            | Queue name for sending reports, use if the service bus is a queue |
| `REPORTER_TOPIC` *             | Yes *    | None            | Topic name for sending reports, use if the service bus is a topic |

> `*` **Notes:** Only set `REPORTER_QUEUE` or `REPORTER_TOPIC` depending on the type of the service bus you are connecting to.

## Event Publish/Subscribe Configs

### Local File System Event Directory

| Variable Name         | Required | Default Value      | Description                                       |
|-----------------------|----------|--------------------|---------------------------------------------------|
| `LOCAL_EVENTS_FOLDER` | No       | `./uploads/events` | Relative file system path to the events directory |

### Azure Event Topics

#### Azure Event Publisher Topic

| Variable Name                 | Required | Default Value | Description                                                           |
|-------------------------------|----------|---------------|-----------------------------------------------------------------------|
| `PUBLISHER_CONNECTION_STRING` | Yes      | None          | Azure connection string with credentials to the event publisher topic |
| `PUBLISHER_TOPIC`             | Yes      | None          | Topic name for the publisher service bus                              |

#### Azure Event Subscriber Subscription

| Variable Name                  | Required | Default Value | Description                                                        |
|--------------------------------|----------|---------------|--------------------------------------------------------------------|
| `SUBSCRIBER_CONNECTION_STRING` | Yes      | None          | Azure connection string with credentials to the event subscription |
| `SUBSCRIBER_TOPIC`             | Yes      | None          | Topic name to subscribe to for receiving events                    |
| `SUBSCRIBER_SUBSCRIPTION`      | Yes      | None          | Subscription name for for the event subscriber                     |

## File Delivery Target Configs

### EDAV Delivery Target

#### Local File System EDAV Directory

| Variable Name       | Required | Default Value    | Description                                            |
|---------------------|----------|------------------|--------------------------------------------------------|
| `LOCAL_EDAV_FOLDER` | No       | `./uploads/edav` | Relative file system path to the EDAV target directory |

#### Azure EDAV Container

| Variable Name                    | Required | Default Value     | Description                                                         |
|----------------------------------|----------|-------------------|---------------------------------------------------------------------|
| `EDAV_STORAGE_ACCOUNT`           | Yes      | None              | Azure EDAV delivery storage account name                            |
| `EDAV_STORAGE_KEY`               | Yes      | None              | Azure EDAV delivery storage account private access key or SAS token |
| `EDAV_ENDPOINT`                  | Yes      | None              | Azure EDAV delivery storage endpoint URL                            |
| `EDAV_CHECKPOINT_CONTAINER_NAME` | No       | `edav-checkpoint` | Container name for EDAV delivery storage checkpoint data            |
| `EDAV_TENANT_ID`                 | No       | None              | EDAV delivery account service principal tenant id                   |
| `EDAV_CLIENT_ID`                 | No       | None              | EDAV delivery account service principal client id                   |
| `EDAV_CLIENT_SECRET`             | No       | None              | EDAV delivery account service principal client secret               |

#### S3 EDAV Bucket

Uses the same authentication defined for the [S3 bucket](#s3-storage-configs)

| Variable Name         | Required | Default Value | Description                                                                                     |
|-----------------------|----------|---------------|-------------------------------------------------------------------------------------------------|
| `EDAV_S3_ENDPOINT`    | Yes      | None          | S3-compatible storage endpoint URL for EDAV delivery storage, must start with `http` or `https` |
| `EDAV_S3_BUCKET_NAME` | Yes      | None          | Bucket name for EDAV delivery storage                                                           |

### EHDI Delivery Target

#### Local File System EHDI Directory

| Variable Name       | Required | Default Value    | Description                                            |
|---------------------|----------|------------------|--------------------------------------------------------|
| `LOCAL_EHDI_FOLDER` | No       | `./uploads/ehdi` | Relative file system path to the EHDI target directory |

#### Azure EHDI Container

| Variable Name                    | Required | Default Value     | Description                                                         |
|----------------------------------|----------|-------------------|---------------------------------------------------------------------|
| `EHDI_STORAGE_ACCOUNT`           | Yes      | None              | Azure EHDI delivery storage account name                            |
| `EHDI_STORAGE_KEY`               | Yes      | None              | Azure EHDI delivery storage account private access key or SAS token |
| `EHDI_ENDPOINT`                  | Yes      | None              | Azure EHDI delivery storage endpoint URL                            |
| `EHDI_CHECKPOINT_CONTAINER_NAME` | No       | `ehdi-checkpoint` | Container name for EHDI delivery storage checkpoint data            |
| `EHDI_TENANT_ID`                 | No       | None              | EHDI delivery account service principal tenant id                   |
| `EHDI_CLIENT_ID`                 | No       | None              | EHDI delivery account service principal client id                   |
| `EHDI_CLIENT_SECRET`             | No       | None              | EHDI delivery account service principal client secret               |

#### S3 EHDI Bucket

Uses the same authentication defined for the [S3 bucket](#s3-storage-configs)

| Variable Name         | Required | Default Value | Description                                                                                     |
|-----------------------|----------|---------------|-------------------------------------------------------------------------------------------------|
| `EHDI_S3_ENDPOINT`    | Yes      | None          | s3-compatible storage endpoint URL for EHDI delivery storage, must start with `http` or `https` |
| `EHDI_S3_BUCKET_NAME` | Yes      | None          | Bucket name for EHDI delivery storage                                                           |

### NCIRD Delivery Target

#### Local File System NCIRD Directory

| Variable Name        | Required | Default Value     | Description                                             |
|----------------------|----------|-------------------|---------------------------------------------------------|
| `LOCAL_NCIRD_FOLDER` | No       | `./uploads/ncird` | Relative file system path to the NCIRD target directory |

#### Azure NCIRD Container

| Variable Name                     | Required | Default Value      | Description                                                          |
|-----------------------------------|----------|--------------------|----------------------------------------------------------------------|
| `NCIRD_STORAGE_ACCOUNT`           | Yes      | None               | Azure NCIRD delivery storage account name                            |
| `NCIRD_STORAGE_KEY`               | Yes      | None               | Azure NCIRD delivery storage account private access key or SAS token |
| `NCIRD_ENDPOINT`                  | Yes      | None               | Azure NCIRD delivery storage endpoint URL                            |
| `NCIRD_CHECKPOINT_CONTAINER_NAME` | No       | `ncird-checkpoint` | Container name for NCIRD delivery storage checkpoint data            |
| `NCIRD_TENANT_ID`                 | No       | None               | NCIRD delivery account service principal tenant id                   |
| `NCIRD_CLIENT_ID`                 | No       | None               | NCIRD delivery account service principal client id                   |
| `NCIRD_CLIENT_SECRET`             | No       | None               | NCIRD delivery account service principal client secret               |

#### S3 NCIRD Bucket

Uses the same authentication defined for the [S3 bucket](#s3-storage-configs)

| Variable Name          | Required | Default Value | Description                                                                                      |
|------------------------|----------|---------------|--------------------------------------------------------------------------------------------------|
| `NCIRD_S3_ENDPOINT`    | Yes      | None          | S3-compatible storage endpoint URL for NCIRD delivery storage, must start with `http` or `https` |
| `NCIRD_S3_BUCKET_NAME` | Yes      | None          | Bucket name for NCIRD delivery storage                                                           |
