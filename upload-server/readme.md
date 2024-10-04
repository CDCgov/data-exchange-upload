# DEX TUSD Go Server

A resumable file upload server for OCIO Data Exchange (DEX)

## Overview

The DEX Upload API is an open-source tool that allows users to upload and manage data sets for public health initiatives. It is designed for ease of use and customization while also ensuring compliance with federal standards.

The DEX Upload API is built on the [tus](https://tus.io) open protocol, which provides a way to upload large files in smaller chunks over HTTP. Each file chunk is sent as an HTTP PATCH request. The last PATCH request tells tus to combine the chunks into the full file. Once the file upload is complete, the server can also push it out to other destinations.

### Key Features

- Resumable uploads
- Support for multiple formats (JSON, XML)
- Error handling and status codes
- Detailed documentation for seamless integration

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

Set the `CSRF_TOKEN` environment variable before starting the server ([see Configuration](#configurations)). You can generate a random 32-byte string for it [here](https://generate-random.org/encryption-key-generator?count=1&bytes=32&cipher=aes-256-cbc&string=&password=)

Running the server with the default configurations will use the local file system for the storage backend and the delivery targets. It will start the Upload API server at [http://localhost:8080](http://localhost:8080) and a Tus client at [http://localhost:8081/](http://localhost:8081/), which will allow you to upload files to the Upload API server.

These files will be uploaded by default to the `upload-server/uploads/tus-prefix/` directory. The base `uploads/` directory name and location can be changed using the `LOCAL_FOLDER_UPLOAD_TUS` environment variable. The `tus-prefix/` name can be changed using the `TUS_UPLOAD_PREFIX` environment variable.

Information about file uploads and delivery are stored as reports and events in the `upload-server/uploads/reports/` and `upload-server/uploads/events` respectively. If delivery targets are specified in the metadata, they will be sent to the corresponding directory in `upload-server/uploads/`.

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

### Configuring Common Service

*upload-server/.env*:

```vim
# common
LOGGER_DEBUG_ON= # set whether the logging level should be DEBUG or INFO
ENVIRONMENT= # set the environment of the service, default=DEV

# server
SERVER_PROTOCOL= # set if the server is http or https, default=http
SERVER_HOSTNAME= # set the hostname of the server, default=localhost
SERVER_PORT= # set the port for the server, default=8080
TUSD_HANDLER_BASE_PATH= # set the path for the upload endpoint, default=/files/
TUSD_HANDLER_INFO_PATH= # set the path for the info endpoint, default=/info/
UPLOAD_CONFIG_PATH= # set the path to the `upload-configs/` directory, default=../upload-configs
EVENT_MAX_RETRY_COUNT= # set the number of retries to publish an event, default=3
METRICS_LABELS_FROM_MANIFEST= # which manifest keys to count in the metrics

# tusd
TUS_UPLOAD_PREFIX= # set sub directory to drop files into within the storage backend, default=tus-prefix

# ui
UI_PORT= # set the port for the UI client, default=8081
CSRF_TOKEN= # set 32-byte string for generating CSRF tokens, this is REQUIRED

# processing status health
PROCESSING_STATUS_HEALTH_URI= # set the URI of the Processing Status Health server
```

### Configuring Distributed File Locking with Redis

When you want to scale this service horizontally, you'll need to use a distributed file locking mechanism to prevent upload corruption. You can read more about the limitations of Tus's support for concurrent requests [here](https://tus.github.io/tusd/advanced-topics/locks/). This service comes with a Redis implementation of a distributed file lock out of the box. All you need is a Redis instance that is accessible from the servers you will deploy this service to. This is provided for you in the [docker-compose set up](#running-using-docker-compose). After the Redis instances is set up, set the following environment variable to enable the use of your Redis instance:

*upload-server/.env*:

```vim
REDIS_CONNECTION_STRING=redis://redispw@cache:6379 # set the URI of the Redis instance
```

>Note: The full URI of the Redis instance must include authentication credentials such as username/password or access token. Make sure to use `rediss://` instead of `redis://` to use TLS for this traffic.

### Configuring OAuth Token Verification Middleware

The Upload API has OAuth token verification middleware for the `/files/` and `/info` endpoints. You can read more about it [here](internal/readme.md). OAuth is disabled by default. If you would like to enable it, you need to set the following environment variables with your OAuth settings:

*upload-server/.env*:

```vim
OAUTH_AUTH_ENABLED=true            # enable or disable OAuth token verification
OAUTH_ISSUER_URL=https://issuer.url # set the URL of the token issuer
OAUTH_REQUIRED_SCOPES="scope1 scope2" # set the space-separated list of required scopes
OAUTH_INTROSPECTION_URL=https://introspection.url # set the introspection url, used for opaque tokens
```

### Configuring the storage backend

This service currently supports local file system, Azure, and AWS as storage backends. You can only use one storage backend at a time. If the Azure configurations are set, it will be the storage backend regardless. If the Azure configurations are not set and the S3 configurations are set, S3 will be the storage backend. If neither Azure or S3 configurations are set, local storage will be the storage backend.

Here are some examples for configuring the different storage backends that this service supports. All of the following commands should be run from the `upload-server/` directory.

#### Local file system

By default, this service uses the file system of the host machine it is running on as a storage backend. Therefore, no environment variables are necessary to set. You can change the directory the service will use as the base of the uploads

*upload-server/.env*:

```vim
LOCAL_FOLDER_UPLOADS_TUS= # set the relative path to the folder where tus will upload files to, default=./uploads
```

#### Azure Storage Account

To upload to an Azure Storage Account, you'll need to collect the name, access key, and endpoint URI of the account. You also need to create a Blob Container within the account. You must set the following environment variables to use Azure.

*upload-server/.env*:

```vim
AZURE_STORAGE_ACCOUNT= # set the name of the storage account
AZURE_STORAGE_KEY= # set the private access key or SAS token of the storage account
AZURE_ENDPOINT= # set the URI of the storage account
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
AZURE_TENANT_ID= # set the tenant id
AZURE_CLIENT_ID= # set the client id
AZURE_CLIENT_SECRET= # set the client secret
```

##### Optional Azure configurations

You can also configure:

*upload-server/.env*:

```vim
TUS_AZURE_CONTAINER_NAME= # set the container name for the uploads blob
DEX_MANIFEST_CONFIG_CONTAINER_NAME= # set the container name for the manifests blob
```

#### S3

To use an AWS S3 bucket as the storage backend, you'll need to create a bucket to upload to within S3 and give a user or service read and write access to it. Then set the bucket name and endpoint URI of the S3 instance.

*upload-server/.env*:

```vim
S3_ENDPOINT= # set the URI of the S3 instance, must start with `http` or `https`
S3_BUCKET_NAME= # set the globally unique name of the S3 bucket
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
AWS_ACCESS_KEY_ID= # set the username or user ID of the user or service account with read and write access to the bucket
AWS_SECRET_ACCESS_KEY= # set the password or private key of a user or service account
AWS_SESSION_TOKEN= # optionally, set the session token for authentication, typically used for short lived keys
AWS_REGION= # set the region of the S3 bucket
```

or using a `profile` in an [AWS Credential file](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html)

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

##### Optional AWS configurations

You can also configure:

*upload-server/.env*:

```vim
DEX_MANIFEST_CONFIG_BUCKET_NAME= # set the name of the bucket containing the `upload-config` files, if not set it defaults to the main bucket
DEX_S3_MANIFEST_CONFIG_FOLDER_NAME= # set the name of the directory in side the bucket containing the `upload-config` files
```

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
