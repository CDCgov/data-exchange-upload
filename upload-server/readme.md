# DEX TUSD Go Server

A resumable file upload server for OCIO Data Exchange (DEX)

## Overview

The DEX Upload API is an open-source tool that allows users to upload and manage data sets for public health initiatives. It is designed for ease of use and customization while also ensuring compliance with federal standards.

The DEX Upload API is built on the [tus](https://tus.io) open protocol, which provides a way to upload large files in smaller chunks over HTTP. Each file chunk is sent as an HTTP PATCH request. The last PATCH request tells tus to combine the chunks into the full file. Once the file upload is complete, the server can also push it out to other destinations.

### Key Features

- Resumable uploads via [tus](https://tus.io)
- File metadata validation
- File routing
- Upload multiple files in parallel
- Configurable authN/authZ middleware
- Support for distributed file locking to enable horizontal scaling
- User authentication and scope enforcement with JWTs

## Folder Structure

Repo is structured (as feasible) based on the [golang-standards/project-layout](https://github.com/golang-standards/project-layout)

## References

- Based on the [tus](https://tus.io/) open protocol for resumable file uploads
- Based on the [tusd](https://github.com/tus/tusd) official reference implementation

---

## Getting Started

### 1. Install Required Tools

Install [Go](https://go.dev/doc/) and/or a container tool (e.g., [Docker](https://www.docker.com/), [Podman](https://podman.io/))

### 2. Clone the Repo

Clone the repo and change into the `upload-server/` inside the repo

```shell
git clone git@github.com:CDCgov/data-exchange-upload.git 

cd data-exchange-upload/upload-server
```

### 3. Start the Server

Running the server with the default configurations will use the local file system for the storage backend and the delivery targets. It will start the Upload API server at [http://localhost:8080](http://localhost:8080) and a Tus client at [http://localhost:8081/](http://localhost:8081/), which will allow you to upload files to the Upload API server.

These files will be uploaded by default to the `upload-server/uploads/tus-prefix/` directory. The base `uploads/` directory name and location can be changed using the `LOCAL_FOLDER_UPLOAD_TUS` environment variable. The `tus-prefix/` name can be changed using the `TUS_UPLOAD_PREFIX` environment variable.

Information about file uploads and delivery are stored as reports and events in the `upload-server/uploads/reports/` and `upload-server/uploads/events` respectively. If delivery targets are specified in the sender manifest config file, they will be sent to the corresponding directory in `upload-server/uploads/`.

Default folder structure

```folder
|-- upload-server
    |-- uploads
        |-- edav // edav delivery destination
        |-- ehdi // ehdi delivery destination
        |-- eicr // eicr delivery destination
        |-- events // file upload and delivery events
        |-- ncird // ncird delivery destination
        |-- reports // file upload and delivery reports
        |-- routing // routing delivery destination
        |-- tus-prefix // file upload destination
```

All of the following commands should be run from the `upload-server/` directory.

#### Flags

##### Application Configuration

`-appconf` passes in an environment variable file to use for configuration

The following forms are permitted:

```shell
-appconf=.env
-appconf .env
--appconf=.env
--appconf .env
```

#### Using Go

##### Running from the Code

Run the code

```shell
go run ./cmd/main.go
```

Run the code with the flag

```shell
go run ./cmd/main.go -appconf=.env 
```

##### Running from the Binary

Build the binary

```shell
go build -o ./dextusd ./cmd/main.go
```

Run the binary

```shell
./dextusd
```

Run the binary with the flag

```shell
./dextusd -appconf=.env
```

#### Using Docker

##### Running Using the Dockerfile

Build the image

```shell
docker build -t dextusdimage .
```

Run the container

```shell
docker run -d -p 8080:8080 -p 8081:8081 --name dextusd dextusdimage
```

> Note: `-p 8080:8080 -p 8081:8081` must be included so that you can access the endpoints

Run the container with the flag, passing in the .env in the same directory

```shell
docker run -d -p 8080:8080 -p 8081:8081 -v .:/conf --name dextusd dextusdimage -appconf=/conf/.env
```

> Note: `-v .:/conf` mounts a volume from your current directory to a directory in the container, which can be anything except `/app` because that is used for the binary. The location of the `-appconf` flag needs to point to this mounted volume. ([see Mount Volume](https://docs.docker.com/reference/cli/docker/container/run/#volume))

##### Running Using Docker Compose

This is the easiest way to start the service locally, because in addition to starting the service it also starts the Redis cache, Prometheus, and Grafana. If there is an `.env` file in the same directory as the docker-compose.yml file, it will automatically use those values when building and starting the containers and it will use the `-appconf` flag to pass the file into the service.

Start the containers using

```shell
docker-compose up -d
```

## Configurations

Configuration of the `upload-server` is managed through environment variables. These environment variables can be set directly in the terminal

(Mac or Linux)

```shell
export SERVER_PORT=8082
go run ./cmd/main.go
```

(Windows)

```shell
set SERVER_PORT=8082
go run ./cmd/main.go
```

or you can create a file and set them in it

*upload-server/env-file*:

```vim
SERVER_PORT=8082
```

then pass the file in using the [`-appconf` flag](#flags)

```shell
go run ./cmd/main.go -appconf env-file
```

or load it into the session using the `source` command

```shell
source env-file
go run ./cmd/main.go
```

If you name this file `.env` you can get the benefits of the [dotenv file format](https://www.dotenv.org/docs/security/env). For instance, it will automatically be recognized and loaded by tools like `docker-compose`. The `.env.example` file in the `upload-server/` directory contains all of the available environment variables for configuring the system. Add any environment variables you would like to set to your `.env` file.

>[!WARNING]
> Never check your `.env` file into source control. It should only be on your local computer or on the server you are using it on.

### Configuration Documentation

Please see [env-configs](./docs/env-configs.md) documentation for more complete documentation on environment configs to be used in the `.env` file.

### Common Service Configurations

*upload-server/.env*:

```vim
## logging and environment
# enable or disable DEBUG logging level, default=INFO
LOGGER_DEBUG_ON= 
# environment for the service, system default=DEV
ENVIRONMENT= 

## server configs
# Protocol use by the server (http or https), default=http
SERVER_PROTOCOL= 
# hostname of the server, default=localhost
SERVER_HOSTNAME= 
# port on which the server runs, default=8080
SERVER_PORT= 
# url path for handling tusd upload requests, default=/files/
TUSD_HANDLER_BASE_PATH= 
# url path for handling tusd info requests, default=/info/
TUSD_HANDLER_INFO_PATH=
# maximum number of retries for event processing, default=3 
EVENT_MAX_RETRY_COUNT= 
# string separated list of keys from the sender manifest config to count in the metrics, default=data_stream_id,data_stream_route,sender_id
METRICS_LABELS_FROM_MANIFEST= 

# tusd
# relative file system path to the tus uploads directory within the storage backend location, default=tus-prefix
TUS_UPLOAD_PREFIX= 

# ui
# port on which the UI client runs, default=8081
UI_PORT= 
# CSRF token used for form security (32 byte string), default=1qQBJumxRABFBLvaz5PSXBcXLE84viE42x4Aev359DvLSvzjbXSme3whhFkESatW
CSRF_TOKEN= 
```

> [!WARNING]
> The default `CSRF_TOKEN` is for development purposes only. You should replace this with a new string, you can generate a 32 byte string [here](https://generate-random.org/encryption-key-generator?count=1&bytes=32&cipher=aes-256-cbc&string=&password=)

### Configuring Distributed File Locking with Redis

When you want to scale this service horizontally, you'll need to use a distributed file locking mechanism to prevent upload corruption. You can read more about the limitations of Tus's support for concurrent requests [here](https://tus.github.io/tusd/advanced-topics/locks/). This service comes with a Redis implementation of a distributed file lock out of the box. All you need is a Redis instance that is accessible from the servers you will deploy this service to. This is provided for you in the [docker-compose set up](#running-using-docker-compose). After the Redis instances is set up, set the following environment variable to enable the use of your Redis instance:

*upload-server/.env*:

```vim
# connection string to the Redis instance
REDIS_CONNECTION_STRING=
```

>Note: The full URI of the Redis instance must include authentication credentials such as username/password or access token. Make sure to use `rediss://` instead of `redis://` to use TLS for this traffic.

### Configuring OAuth Token Verification Middleware

The Upload API has OAuth token verification middleware for the `/files/` and `/info` endpoints. You can read more about it [here](internal/readme.md). OAuth is disabled by default. If you would like to enable it, you need to set the following environment variables with your OAuth settings:

*upload-server/.env*:

```vim
# enable or disable OAuth token verification
OAUTH_AUTH_ENABLED=false
# URL of the OAuth token issuer
OAUTH_ISSUER_URL=
# space-separated list of required scopes
OAUTH_REQUIRED_SCOPES=
# optionally, URL for OAuth introspection, used for opaque tokens
OAUTH_INTROSPECTION_URL=
```

### Configuring the storage backend

This service currently supports local file system, Azure, and AWS as storage backends. You can only use one storage backend at a time. If the Azure configurations are set, it will be the storage backend regardless. If the Azure configurations are not set and the S3 configurations are set, S3 will be the storage backend. If neither Azure or S3 configurations are set, local storage will be the storage backend.

#### Local file system

By default, this service uses the file system of the host machine it is running on as a storage backend. Therefore, no environment variables are necessary to set. You can change the directory the service will use as the base of the uploads

*upload-server/.env*:

```vim
# relative file system path to the base upload directory, default=./uploads
LOCAL_FOLDER_UPLOADS_TUS= 
```

##### Sender manifest config location

By default, the service uses the sender manifest config files located in `../upload-configs`. These files within this directory are split into the sub directories `v1` and `v2` depending on their version. The service will default to the v2 files if there are two versions of the same sender manifest config. You can change the base of the configs directory

*upload-server/.env*:

```vim
# relative file system path to the sender manifest configuration directory, default=../upload-configs
UPLOAD_CONFIG_PATH= 
```

#### Azure Storage Account

To upload to an Azure Storage Account, you'll need to collect the name, access key, and endpoint URI of the account. You also need to create a [Blob container](https://learn.microsoft.com/en-us/azure/storage/blobs/quickstart-storage-explorer) within the account. You must set the following environment variables to use Azure.

*upload-server/.env*:

```vim
# Azure storage account name
AZURE_STORAGE_ACCOUNT= 
# Azure storage account private access key or SAS token
AZURE_STORAGE_KEY= 
# Azure storage endpoint URL
AZURE_ENDPOINT= 
# container name for tus base upload storage
TUS_AZURE_CONTAINER_NAME= 
```

##### Azure local development

For local development, you can use [Azurite](https://learn.microsoft.com/en-us/azure/storage/common/storage-use-azurite?tabs=visual-studio,blob-storage#install-azurite) to emulate Azure Storage. There is a Docker Compose file included here, `docker-compose.azurite.yml` that creates and starts an Azurite container. You only need to set the `AZURE_STORAGE_KEY`. You can get [the default Azurite key here](https://github.com/Azure/Azurite?tab=readme-ov-file#default-storage-account). To start the service with an Azure storage backend, run

```shell
podman-compose -f docker-compose.yml -f docker-compose.azurite.yml up -d
```

##### Optional Azure authentication

To use [Azure Service Principal](https://learn.microsoft.com/en-us/entra/identity-platform/app-objects-and-service-principals?WT.mc_id=devops-10986-petender&tabs=browser) for authentication, set:

*upload-server/.env*:

```vim
# Azure storage account service principal tenant id
AZURE_TENANT_ID= 
# Azure storage account service principal client id
AZURE_CLIENT_ID= 
# Azure storage account service principal client secret
AZURE_CLIENT_SECRET= 
```

##### Optional Azure blob for sender manifest config files

If you would like to store the sender manifest config files on Azure, create a blob container for them using the same credentials as the upload blob container. Copy the `../upload-configs/v1` and `../upload-configs/v2` directories to the blob. Set `DEX_MANIFEST_CONFIG_CONTAINER_NAME` to the new blob container name

*upload-server/.env*:

```vim
# container name for sender manifest configuration files
DEX_MANIFEST_CONFIG_CONTAINER_NAME= 
```

If `DEX_MANIFEST_CONFIG_CONTAINER_NAME` is not set, the sender manifest config files on the file system will be used.

#### S3

To use an AWS S3 bucket as the storage backend, you'll need to [create a bucket](https://docs.aws.amazon.com/AmazonS3/latest/userguide/create-bucket-overview.html) to upload to within S3 and give a user or service read and write access to it. Then set the bucket name and endpoint URI of the S3 instance.

*upload-server/.env*:

```vim
# s3-compatible storage endpoint URL, must start with `http` or `https`
S3_ENDPOINT= 
# bucket name for tus base upload storage
S3_BUCKET_NAME= 
```

##### AWS local development

For local development, you can use [Minio](https://min.io/docs/minio/container/index.html) to emulate the AWS S3 API. There is a Docker Compose file included here, `docker-compose.minio.yml` that creates and starts a Minio container. To start the service with an AWS S3 storage backend, run

```shell
podman-compose -f docker-compose.yml -f docker-compose.minio.yml up -d
```

##### AWS Authentication

Authentication is handled using the standard AWS environment variables

*upload-server/.env*:

```vim
# username or user ID of the user or service account with read and write access to the bucket
AWS_ACCESS_KEY_ID= 
# password or private key of a user or service account with read and write access to the bucket
AWS_SECRET_ACCESS_KEY= 
# optional, session token for authentication (typically used for short lived keys)
AWS_SESSION_TOKEN= 
# region of the s3 bucket
AWS_REGION= 
```

or using a `profile` in an [AWS Credential file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) with the AWS CLI

*~/.aws/credentials/credentials*:

```vim
[default]
aws_access_key_id = <YOUR_DEFAULT_ACCESS_KEY_ID>
aws_secret_access_key = <YOUR_DEFAULT_SECRET_ACCESS_KEY>
aws_session_token = <YOUR_SESSION_TOKEN>

[test-account]
aws_access_key_id = <YOUR_TEST_ACCESS_KEY_ID>
aws_secret_access_key = <YOUR_TEST_SECRET_ACCESS_KEY>
```

and set the region in an [AWS Config file](https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/#creating-the-config-file)

*~/.aws/config/config*:

```vim
[default]
region = <REGION>

[profile test-account]
region = <REGION>
```

> Note: If you use a credential profile that is not [default], you need to explicitly set the `AWS_PROFILE` environment variable to the profile you want to use, before starting the service.

##### Optional AWS S3 bucket for sender manifest config files

If you would like to store the sender manifest config files in an AWS S3 bucket, set the `DEX_S3_MANIFEST_CONFIG_FOLDER_NAME` environment variable to the directory in the bucket to use. Optionally, you can also create a new bucket for the configs and set `DEX_MANIFEST_CONFIG_BUCKET_NAME` environment variable to that new bucket name. The new bucket must use the same credentials as the upload bucket. Copy the`../upload-configs/v1` and `../upload-configs/v2` directories to the new config folder in the bucket

*upload-server/.env*:

```vim
# Directory name inside the s3 bucket for the sender manifest configuration files within the upload bucket
DEX_S3_MANIFEST_CONFIG_FOLDER_NAME= 
# Bucket name for the sender manifest configurations, if not set it defaults to the upload bucket
DEX_MANIFEST_CONFIG_BUCKET_NAME= 
```

If neither `DEX_MANIFEST_CONFIG_BUCKET_NAME` or `DEX_S3_MANIFEST_CONFIG_FOLDER_NAME` are set, the sender manifest config files on the file system will be used.

### Configuring the reports location

By default, the reports of upload and delivery activity are written to the `./uploads/reports` directory in the local file system.

#### Local file system reports directory

To change the location of the report files

*upload-server/.env*:

```vim
# relative file system path to the reports directory
LOCAL_REPORTS_FOLDER= 
```

#### Azure report service bus

Create an Azure service bus [queue](https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-quickstart-portal) or [topic](https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-quickstart-topics-subscriptions-portal) for publishing the report messages. Set the following environment variables with the details from the new service bus

*upload-server/.env*:

```vim
# Azure connection string with credential to the queue or topic
REPORTER_CONNECTION_STRING= 
# queue name for sending reports, use if the service bus is a queue
REPORTER_QUEUE= 
# topic name for sending reports, use if the service bus is a queue
REPORTER_TOPIC= 
```

### Configuring the event publication and subscription

By default, the event messages about upload and delivery activity are written to the `./uploads/events` directory in the local file system.

#### Local file system events directory

To change the location of the event files

*upload-server/.env*:

```vim
# relative file system path to the events directory
LOCAL_EVENTS_FOLDER= 
```

#### Azure event publisher and subscriber service buses

##### Event publisher

Create an Azure service bus [topic](https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-quickstart-topics-subscriptions-portal) for publishing event messages. Set the following environment variables with the details from the new service bus topic

*upload-server/.env*:

```vim
# Azure connection string with credentials to the event publisher topic
PUBLISHER_CONNECTION_STRING= 
# topic name for the event publisher service bus
PUBLISHER_TOPIC= 
```

##### Subscriber

Create a [subscription](https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-quickstart-topics-subscriptions-portal#create-subscriptions-to-the-topic) for the desired Azure service bus topic. Set the environment variables with the details from the new subscription

*upload-server/.env*:

```vim
# Azure connection string with credentials to the event subscription
SUBSCRIBER_CONNECTION_STRING= 
# topic name to subscribe to for receiving events
SUBSCRIBER_TOPIC= 
# subscription name for the event subscriber
SUBSCRIBER_SUBSCRIPTION= 
```

### Configuring upload routing and delivery targets

This service is capable of copying files that are uploaded to other storage locations, even ones that are outside the on-prem or cloud environment your service is deployed to.  This is useful when you want your files to land in particular storage locations based on their metadata.  Setting this up begins with the creation of a
YML file that defines delivery groups, and one or more delivery targets.  These targets currently support Azure Blob, S3, and local file system.

By default, this service will use the YML file located at `configs/local/delivery.yml`, but you can create your own and point to it via the `DEX_DELIVERY_CONFIG_FILE` environment variable.

Start by defining programs, which act as delivery groups

*configs/local/delivery.yml*:

```yml
programs:
  - data_stream_id: teststream1
    data_stream_route: testroute1
  - data_stream_id: teststream2
    data_stream_route: testroute2
```

Next, define at least one delivery target for each group.  Each of these target endpoints can be configured independently to point to a local file system directory, an Azure Blob container, or an AWS S3 bucket. Specify the type of connection you want by setting the `type` field to either `az-blob`, `s3`, or `file`.

#### Azure blob target

Create an Azure [Blob container](https://learn.microsoft.com/en-us/azure/storage/blobs/quickstart-storage-explorer) or get the values from an existing Blob container. Then, set the required connection information, which can be a SAS token and connection string, or Azure service principle.  *Note that the service will create the container if it does not already exist*.

*configs/local/delivery.yml*:

```yml
programs:
  - data_stream_id: teststream1
    data_stream_route: testroute1
    delivery_targets:
      - name: target1
        type: az-blob
        endpoint: https://target1.blob.core.windows.net
        container_name: target1_container
        tenant_id: $AZURE_TENANT_ID
        client_id: $AZURE_CLIENT_ID
        client_secret: $AZURE_CLIENT_SECRET
  - data_stream_id: teststream2
    data_stream_route: testroute2
    delivery_targets:
      - name: target2
        type: az-blob
        endpoint: https://target2.blob.core.windows.net
        container_name: target2_container
        tenant_id: $AZURE_TENANT_ID
        client_id: $AZURE_CLIENT_ID
        client_secret: $AZURE_CLIENT_SECRET
```

*Note that you can substiture environment variables using the `$` notation.  This is so you can keep secrets like service principle credentials or SAS tokens out of this configuration file.*

#### AWS S3 bucket target

Create an AWS [S3 bucket](https://docs.aws.amazon.com/AmazonS3/latest/userguide/create-bucket-overview.html) or get the values from an existing S3 bucket.  Then, set the access credentials and endpoint for the bucket in the following way:

*configs/local/delivery.yml*:

```yml
programs:
  - data_stream_id: teststream1
    data_stream_route: testroute1
    delivery_targets:
      - name: target1
        type: s3
        endpoint: https://target1.s3.aws.com
        bucket_name: target1
        access_key_id: $S3_ACCESS_KEY_ID
        secret_access_key: $S3_SECRET_ACCESS_KEY
        REGION: us-east-1
  - data_stream_id: teststream2
    data_stream_route: testroute2
    delivery_targets:
      - name: target2
        type: s3
        endpoint: https://target2.s3.aws.com
        bucket_name: target2
        access_key_id: $S3_ACCESS_KEY_ID
        secret_access_key: $S3_SECRET_ACCESS_KEY
        REGION: us-east-1
```

#### Local file system target

To use a local file system target, you simply need to set a directory path.  *Note that the service will create the path if it does not exist*

*configs/local/delivery.yml*:

```yml
programs:
  - data_stream_id: teststream1
    data_stream_route: testroute1
    delivery_targets:
      - name: target1
        type: file
        path: /my/uploads/target1
  - data_stream_id: teststream2
    data_stream_route: testroute2
    delivery_targets:
      - name: target2
        type: file
        path: /my/uploads/target2
```

### Configuring Processing Status API Integration

Upload server is capable of being run locally with the [Processing Status API](https://github.com/CDCgov/data-exchange-processing-status) to integrate features from that service into the Upload end to end flow.  Setting this up allows for the capability of integrataing reporting structures into the bigger Upload workflow.  The Processing Status API repository will need to be cloned locally to access its features for integration. This setup currently assumes that the repositories live adjacent to each other on the local filesystem.

#### Build Processing Status Report Sink

PStatus report sink needs to be built so your container system gets an image with the changes. To do this run the following from the `pstatus-report-sink-ktor directory`.

For building the client for Podman:
```
./gradlew jibDockerBuild -Djib.dockerClient.executable=$(which podman)
```

For building the client for Docker:
```
./gradlew jibDockerBuild
```

This will create a local container that can be used by the following steps.

From the `upload-server` directory in the Upload Server repository run the following command to built a system using both Upload and PS API.

> [!NOTE]
> In order to avoid port colissions, set the environment variable `SINK_PORT` to a value other than `8080`.  Suggested port: `8082`

```
podman-compose -f docker-compose.yml -f docker-compose.localstack.yml -f ../../data-exchange-processing-status/docker-compose.yml -f compose.pstatus.yml up -d
```

This will set up the system so that:
- Upload API is available locally on port `8080`
- Upload UI is available locally on port `8081`
- PS API GraphQL endpoint is available locally on `8090`


## Testing

### Unit Tests

> [!TIP]
> Before running unit tests, make sure to clean the file system with the `clean.sh` script. This removes any temporary upload and report files that the tests generated.

Run the unit tests

```go
go test ./...
```

Run the unit tests with code coverage

```go
go test -coverprofile=c.out ./...
go tool cover -html=c.out
```

### Integration Tests (with minio and azurite)

Set the environment variable `AZURE_STORAGE_KEY` in your `.env` file or locally in your terminal. You can get [the default key here](https://github.com/Azure/Azurite?tab=readme-ov-file#default-storage-account).

```shell
podman-compose -f docker-compose.yml -f docker-compose.azurite.yml -f docker-compose.minio.yml -f docker-compose.testing.yml up --exit-code-from upload-server
```

## VS Code

When using VS Code, we recommend using the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) made by the Go Team at Google

*.vscode/launch.json*:

```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "cmd/main.go",
            "cwd": "${workspaceFolder}",
            "args": []
        }
    ]
}
```

## Future Enhancements
### Proper OIDC user authentication for the UI
The current login process for the UI is for the user to input their SAMS JWT that they aquired offline.  Crucually, this created an attack vector where an attacker can log into the UI
of they aquire a leaked JWT.  This is a low risk at the moment because the UI is not exposed publicly, but would otherwise be a serious vulnerability.  It is also a non-ideal user experience as the user should not have to leave their browser session in order to log into the UI.

To enhance this and cover this potential attack vector, the upload service UI should have a proper "Login with SAMS" button on the login page.  Clicking this shall redirect a user to the SAMS login page where they can authenticate and grant SAMS access to the Upload API on their behalf.  This will subsequently redirect the user back to the UI landing page and take care of the credential/JWT exchange under the hood.

It's recommented to implement this enhancement in a way that other identity providers could be easly configured.  Ideally, the Upload API should accept a set of one or more provider information use it to dynamically generate the login buttons on the login page.