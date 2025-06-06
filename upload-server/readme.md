# DEX TUSD Go Server

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

## Getting Started

### 1. Install Required Tools

Install [Go](https://go.dev/doc/) and/or a container tool (e.g., [Docker](https://www.docker.com/), [Podman](https://podman.io/))

### 2. Clone the Repo

Clone the repo and navigate to the `upload-server/` directory:

```shell
git clone git@github.com:CDCgov/data-exchange-upload.git 

cd data-exchange-upload/upload-server
```

### 3. Start the Server

Running the server with the default configurations will use the local file system for the storage backend and the delivery targets. It will start the Upload API server at [http://localhost:8080](http://localhost:8080) and a Tus client at [http://localhost:8081/](http://localhost:8081/), which will allow you to upload files to the Upload API server.

Files will be uploaded by default to the `upload-server/uploads/tus-prefix/` directory. The base `uploads/` directory name and location can be changed using the `LOCAL_FOLDER_UPLOAD_TUS` environment variable. The `tus-prefix/` name can be changed using the `TUS_UPLOAD_PREFIX` environment variable.

Information about file uploads and delivery are stored as reports and events in `upload-server/uploads/reports/` and `upload-server/uploads/events` directory paths, respectively. If delivery targets are specified in the sender manifest config file, they will be sent to the corresponding directory in `upload-server/uploads/`.

*Default folder structure*

```folder
|-- upload-server
    |-- uploads
        |-- edav // edav delivery destination
        |-- events // file upload and delivery events
        |-- reports // file upload and delivery reports
        |-- routing // routing delivery destination
        |-- tus-prefix // file upload destination
```

All of the following commands should be run from the `upload-server/` directory.

#### Flags

##### Application Configuration

The command `-appconf` passes in an environment variable file to use for configuration.

The following forms are permitted:

```shell
-appconf=.env
-appconf .env
--appconf=.env
--appconf .env
```

#### Using Go

##### Running from the Code

Run the code:

```shell
go run ./cmd/main.go
```

Run the code with the flag:

```shell
go run ./cmd/main.go -appconf=.env 
```

##### Running from the Binary

Build the binary:

```shell
go build -o ./dextusd ./cmd/main.go
```

Run the binary:

```shell
./dextusd
```

Run the binary with the flag:

```shell
./dextusd -appconf=.env
```

#### Using Docker

##### Running Using the Dockerfile

Build the image:

```shell
docker build -t dextusdimage .
```

Run the container:

```shell
docker run -d -p 8080:8080 -p 8081:8081 --name dextusd dextusdimage
```

> Note: Include `-p 8080:8080 -p 8081:8081` to ensure access to the endpoints.

Run the container with the flag, passing in the .env in the same directory:

```shell
docker run -d -p 8080:8080 -p 8081:8081 -v .:/conf --name dextusd dextusdimage -appconf=/conf/.env
```

> Note: `-v .:/conf` mounts a volume from the current directory to a directory in the container, which can be anything except `/app` because that is used for the binary. The location of the `-appconf` flag needs to point to this mounted volume. ([see Mount Volume](https://docs.docker.com/reference/cli/docker/container/run/#volume))

##### Running Using Docker Compose

This is the easiest way to start the service locally, because in addition to starting the service it also starts the Redis cache, Prometheus, and Grafana. If there is an `.env` file in the same directory as the docker-compose.yml file, it will automatically use those values when building and starting the containers and it will use the `-appconf` flag to pass the file into the service.

Start the containers using:

```shell
docker-compose up -d
```

## Configurations

Configuration of the `upload-server` is managed through environment variables. 

These environment variables can be set directly in the terminal:

*(Mac or Linux)*

```shell
export SERVER_PORT=8082
go run ./cmd/main.go
```

*(Windows)*

```shell
set SERVER_PORT=8082
go run ./cmd/main.go
```

or set within a newly created file:

*upload-server/env-file*:

```vim
SERVER_PORT=8082
```

then pass in the file using the [`-appconf` flag](#flags):

```shell
go run ./cmd/main.go -appconf env-file
```

or load it into the session using the `source` command:

```shell
source env-file
go run ./cmd/main.go
```

If this file is named `.env` the benefits of the [dotenv file format](https://www.dotenv.org/docs/security/env) can be leveraged. For instance, it will automatically be recognized and loaded by tools like `docker-compose`. The `.env.example` file in the `upload-server/` directory contains all of the available environment variables for configuring the system. Additional custom environment variables can also be included in the `.env` file, as needed.

>[!WARNING]
> Never check a personal `.env` file into source control. It should only reside locally or on the server it is used on.

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
> The default `CSRF_TOKEN` is for development purposes only and should be replaced with a new string; generate a 32 byte string [here](https://generate-random.org/encryption-key-generator?count=1&bytes=32&cipher=aes-256-cbc&string=&password=)

### Configuring Distributed File Locking with Redis

To scale this service horizontally, utilization of a distributed file locking mechanism to prevent upload corruption is needed. More information about the limitations of Tus' support for concurrent requests [here](https://tus.github.io/tusd/advanced-topics/locks/). This service comes with a Redis implementation of a distributed file lock. All that is needed is a Redis instance that is accessible from the servers to which this service will be deployed. This is provided in the [docker-compose set up](#running-using-docker-compose).

After the Redis instances are set up, set the following environment variable to enable use:

*upload-server/.env*:

```vim
# connection string to the Redis instance
REDIS_CONNECTION_STRING=
```

>Note: The full URI of the Redis instance must include authentication credentials such as username/password or access token. Make sure to use `rediss://` instead of `redis://` to use TLS for this traffic.

### Configuring OAuth Token Verification Middleware

The Upload API has OAuth token verification middleware for the `/files/` and `/info` endpoints. You can read more about it [here](internal/readme.md). OAuth is disabled by default. 

To enable it, set the following environment variables with your OAuth settings:

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

This service currently supports local file system, Azure, and AWS as storage backends. Only one storage backend can be used at a time. If the Azure configurations are set, it will be the storage backend regardless. If the Azure configurations are not set and the S3 configurations are set, S3 will be the storage backend. If neither Azure or S3 configurations are set, local storage will be the storage backend.

#### Local file system

By default, this service uses the file system of the host machine it is running on as a storage backend. Therefore, no environment variables are necessary to set. The directory the service will use as the base of the uploads can be changed, if needed.

*upload-server/.env*:

```vim
# relative file system path to the base upload directory, default=./uploads
LOCAL_FOLDER_UPLOADS_TUS= 
```

##### Sender manifest config location

By default, the service uses the sender manifest config files located in `../upload-configs`.

*upload-server/.env*:

```vim
# relative file system path to the sender manifest configuration directory, default=../upload-configs
UPLOAD_CONFIG_PATH= 
```

#### Azure Storage Account

To upload to an Azure storage account, collect the name, access key, and endpoint URI and create a [Blob container](https://learn.microsoft.com/en-us/azure/storage/blobs/quickstart-storage-explorer) within the storage account. 

Set the following environment variables to use Azure.

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

For local development, [Azurite](https://learn.microsoft.com/en-us/azure/storage/common/storage-use-azurite?tabs=visual-studio,blob-storage#install-azurite) can be used to emulate Azure Storage. There is a Docker Compose file included here, `docker-compose.azurite.yml` that creates and starts an Azurite container. Set the `AZURE_STORAGE_KEY` and get the [default Azurite key here](https://github.com/Azure/Azurite?tab=readme-ov-file#default-storage-account). 

To start the service with an Azure storage backend, run:

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

To store the sender manifest config files on Azure, create a blob container for them using the same credentials as the upload blob container. Copy the `../upload-configs` directory to the blob. 

Set `DEX_MANIFEST_CONFIG_CONTAINER_NAME` to the new blob container name.

*upload-server/.env*:

```vim
# container name for sender manifest configuration files
DEX_MANIFEST_CONFIG_CONTAINER_NAME= 
```

If `DEX_MANIFEST_CONFIG_CONTAINER_NAME` is not set, the sender manifest config files on the file system will be used.

#### S3

To use an AWS S3 bucket as the storage backend, [create a bucket](https://docs.aws.amazon.com/AmazonS3/latest/userguide/create-bucket-overview.html) to upload to within S3 and give a user or service read and write access to it. Then set the bucket name and endpoint URI of the S3 instance.

*upload-server/.env*:

```vim
# s3-compatible storage endpoint URL, must start with `http` or `https`
S3_ENDPOINT= 
# bucket name for tus base upload storage
S3_BUCKET_NAME= 
```

##### AWS local development

For local development, [Minio](https://min.io/docs/minio/container/index.html) can be used to emulate the AWS S3 API. There is a Docker Compose file included here, `docker-compose.minio.yml` that creates and starts a Minio container. 

To start the service with an AWS S3 storage backend, run:

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

> Note: If a credential profile is used that is not [default], the `AWS_PROFILE` environment variable must be explicitly set to the desired profile to use, before starting the service.

##### Optional AWS S3 bucket for sender manifest config files

If the sender manifest config files are to be stored in an AWS S3 bucket, set the `DEX_S3_MANIFEST_CONFIG_FOLDER_NAME` environment variable to the directory in the bucket to use. Optionally, a new bucket can be created for the configs, setting `DEX_MANIFEST_CONFIG_BUCKET_NAME` environment variable to that new bucket name. The new bucket must use the same credentials as the upload bucket. 

Copy the `../upload-configs` directory to the new config folder in the bucket

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

To change the location of the report files:

*upload-server/.env*:

```vim
# relative file system path to the reports directory
LOCAL_REPORTS_FOLDER= 
```

#### Azure report service bus

Create an Azure service bus [queue](https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-quickstart-portal) or [topic](https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-quickstart-topics-subscriptions-portal) for publishing the report messages. 

Set the following environment variables with the details from the new service bus:

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

To change the location of the event files:

*upload-server/.env*:

```vim
# relative file system path to the events directory
LOCAL_EVENTS_FOLDER= 
```

#### Azure event publisher and subscriber service buses

##### Event publisher

Create an Azure service bus [topic](https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-quickstart-topics-subscriptions-portal) for publishing event messages. 

Set the following environment variables with the details from the new service bus topic:

*upload-server/.env*:

```vim
# Azure connection string with credentials to the event publisher topic
PUBLISHER_CONNECTION_STRING= 
# topic name for the event publisher service bus
PUBLISHER_TOPIC= 
```

##### Subscriber

Create a [subscription](https://learn.microsoft.com/en-us/azure/service-bus-messaging/service-bus-quickstart-topics-subscriptions-portal#create-subscriptions-to-the-topic) for the desired Azure service bus topic. 

Set the environment variables with the details from the new subscription:

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

This service is capable of copying files that are uploaded to other storage locations, even those outside the on-prem or cloud environment to which the service is deployed. This is useful when file delivery locations are configurable based on provided metadata. Setting this up begins with the creation of a YML file that defines delivery groups, and one or more delivery targets. These targets currently support Azure Blob, S3, and local file system.

By default, this service will use the YML file located at `configs/local/delivery.yml`, but a custom delivery file can be created and configured via the `DEX_DELIVERY_CONFIG_FILE` environment variable.

Start by defining programs, which act as delivery groups.

*configs/local/delivery.yml*:

```yml
programs:
  - data_stream_id: teststream1
    data_stream_route: testroute1
  - data_stream_id: teststream2
    data_stream_route: testroute2
```

Next, define at least one delivery target for each group. Each of these target endpoints can be configured independently to point to a local file system directory, an Azure Blob container, or an AWS S3 bucket. Specify the type of connection by setting the `type` field to either `az-blob`, `s3`, or `file`.

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

*Note that environment variables can be substituted using the `$` notation. This facilitates keeping secrets like service principle credentials or SAS tokens out of this configuration file.*

#### AWS S3 bucket target

Create an AWS [S3 bucket](https://docs.aws.amazon.com/AmazonS3/latest/userguide/create-bucket-overview.html) or get the values from an existing S3 bucket. 

Then, set the access credentials and endpoint for the bucket in the following way:

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

To use a local file system target, simply set a directory path. *Note that the service will create the path if it does not exist*

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

Upload server is capable of being run locally with the [Processing Status API](https://github.com/CDCgov/data-exchange-processing-status) to integrate features from that service into the Upload end to end flow. Setting this up allows for the capability of integrataing reporting structures into the bigger Upload workflow. The Processing Status API repository will need to be cloned locally to access its features for integration. This setup currently assumes that the repositories live adjacent to each other on the local filesystem.

#### Build Processing Status Report Sink

PStatus report sink needs to be built so your container system gets an image with the changes. 

To do this run the following from the `pstatus-report-sink-ktor directory`:

For building the client for Podman:
```
./gradlew jibDockerBuild -Djib.dockerClient.executable=$(which podman)
```

For building the client for Docker:
```
./gradlew jibDockerBuild
```

This will create a local container that can be used by the following steps.

From the `upload-server` directory in the Upload Server repository run the following command to build a system using both the Upload and Processing Status services.

> [!NOTE]
> In order to avoid port collisions, set the environment variable `SINK_PORT` to a value other than `8080`.  Suggested port: `8082`

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

Run the unit tests:

```go
go test ./...
```

Run the unit tests with code coverage:

```go
go test -coverprofile=c.out ./...
go tool cover -html=c.out
```

### Integration Tests (with minio and azurite)

Set the environment variable `AZURE_STORAGE_KEY` in the `.env` file or locally in the terminal. Get the [default key here](https://github.com/Azure/Azurite?tab=readme-ov-file#default-storage-account).

```shell
podman-compose -f docker-compose.yml -f docker-compose.azurite.yml -f docker-compose.minio.yml -f docker-compose.testing.yml up --exit-code-from upload-server
```

## VS Code

When using VS Code, it is recommended to use the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.Go) made by the Go Team at Google.

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
The current login process for the UI is for the user to input their SAMS JWT that they aquired offline. Crucially, this created an attack vector where an attacker can log into the UI if they aquire a leaked JWT. This is a low risk at the moment because the UI is not exposed publicly, but would otherwise be a serious vulnerability.  It is also a non-ideal user experience as the user should not have to leave their browser session in order to log into the UI.

To enhance this and cover this potential attack vector, the upload service UI should have a proper "Login with SAMS" button on the login page. Clicking this shall redirect a user to the SAMS login page where they can authenticate and grant SAMS access to the Upload API on their behalf. This will subsequently redirect the user back to the UI landing page and take care of the credential/JWT exchange under the hood.

It's recommended to implement this enhancement in a way that other identity providers could be easly configured.  Ideally, the Upload API should accept a set of one or more provider's information and use it to dynamically generate the login buttons on the login page.

### Enhanced error triaging with contextually aware logs
Server logs can be enhanced to include the trace and span IDs for the Tempo trace and span that is capturing the log. This will enable a capability in Grafana that links logs and traces together, and allows the user to easily navigate between a span and the logs produced within it. This will significantly minimize error triage time. If an error occurs during the critical path of an upload, the error log will be surfaced with the error span as opposed to hunting for the log from the error span's time range, or hunting for the span given the error log's timestamp.

It appears that Tempo and Loki can be configured to create the log to trace linking without any code changes.  [https://grafana.com/docs/grafana/latest/datasources/tempo/#trace-to-logs](https://grafana.com/docs/grafana/latest/datasources/tempo/#trace-to-logs). This would be an ideal approach if Loki and Tempo are part of your monitoring stack. Otherwise, code changes will need to be made to compose the server's logger context with the current trace and span IDs.

### Upload delivery performance optimization
File delivery is one of the biggest hits to latency that end users of the Upload API experience. It can take 10s of seconds to transfer a file of only a few hundred megabytes. This gets significantly worse for cross-cloud transfers that go through more network hops. Evidence for this can be seen in traces emitted by the Upload API.

The Azure SDK that the Upload API uses to perform the copy of an upload object from an S3 bucket to an Azure Blob container has the capability of chunking the object and concurrently copying the chunks across without losing file integrity. This takes advantage of the "staging" operation within Azure Blob that allows clients to quickly iterate over the chunks and mark, or stage, them for copy. Finally, the client commits all of the staged chunks and the actual copying of the data is delegated to Azure itself. Work for this has already been done in the following draft PR: [https://github.com/CDCgov/data-exchange-upload/pull/536](https://github.com/CDCgov/data-exchange-upload/pull/536).  

Another approach may be to utilize the parallelization mechanisms within the Azure blob SDK for Golang itself. It appears that the upload operations can take a concurrency parameter that tells the client how to parallelize the chunk uploading. [https://learn.microsoft.com/en-us/azure/storage/blobs/storage-blob-upload-go](https://learn.microsoft.com/en-us/azure/storage/blobs/storage-blob-upload-go).