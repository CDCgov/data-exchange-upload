# DEX TUSD Go Server

A resumable file upload server for OCIO Data Exchange (DEX)

## Folder structure

Repo is structured (as feasible) based on the [golang-standards/project-layout](https://github.com/golang-standards/project-layout)

## References

- Based on the [tus](https://tus.io/) open protocol for resumable file uploads
- Based on the [tusd](https://github.com/tus/tusd) official reference implementation

## Usage

### Configuring the storage backend

This service currently supports local file system, Azure, and AWS as storage backends. Configuring a storage backend
for your setup is done via environment variables. Environment variables can be set at the system level, or via a `.env` file
located within the `configs/local/` directory. Here are some examples for configuring the different storage backends
that this service supports.

#### Local file system

By default, this service uses the file system of the host machine it is running on as a storage backend. Therefore, no
environment variables are necessary to set. To run, simply execute

```go
go run ./cmd/main.go
```

This will start the HTTP server at http://localhost:8080. With a Tus client, you can upload files to http://localhost:8080/files,
and they will show up in the `uploads/` directory.

You can configure this behavior with the following environment variables:

- `SERVER_PORT` - Sets the port that the server runs on. Default is 8080.
- `LOCAL_FOLDER_UPLOADS_TUS` - Relative path to the folder where tus will upload files to. Default is `./uploads`.
- `TUSD_HANDLER_BASE_PATH` - URL path for the file upload endpoint. Default is `/files` or `/files/`.
- `TUS_UPLOAD_PREFIX` - Sub folder to drop files into within the local or cloud folder. Defaults to `tus-prefix`.

#### Azure Storage Account

To upload to an Azure Storage Account, you'll need to collect the name, access key, and endpoint URI of the account. You
also need to create a Blob Container within the account. Next, fill in the following environment variables to tell the
service to use your Azure storage account as the storage backend:

- `AZURE_STORAGE_ACCOUNT` - Name of the storage account.
- `AZURE_STORAGE_KEY` - Private access key or SAS token of the account.
- `AZURE_ENDPOINT` - URI of the storage account.

#### S3

This service currently uses the AWS S3 SDK to support uploading to S3 storage backends, however any compatible S3 storage,
such as Minio, is supported. You'll need to set the following environment variables to enable this:

- `AWS_S3_ENDPOINT` - URI of the S3 instance. Must start with `http` or `https`.
- `AWS_S3_BUCKET_NAME` - Globally unique name of the S3 bucket.
- `AWS_REGION` - Region where the S3 bucket lives.
- `AWS_ACCESS_KEY_ID` - Username or user ID of a user or service account with write access to the bucket.
- `AWS_SECRET_ACCESS_KEY` - Password or private key of user or service account.
- `AWS_SESSION_TOKEN` - Optional session token for authentication. Typically used for short lived keys.

### Configuring Distributed File Locking with Redis

When you want to scale this service horizontally, you'll need to use a distributed file locking mechanism to prevent
upload corruption. You can read more about the limitations of Tus's support for concurrent requests [here](https://tus.github.io/tusd/advanced-topics/locks/).
This service comes with a Redis implementation of a distributed file lock out of the box. All you need is a Redis instance
that is accessible from the servers you will deploy this service to. After that, set the following environment variable
to enable the use of your Redis instance:

- `REDIS_CONNECTION_STRING` - The full URI of the Redis instance, including authentication credentials such as username/password or access token. Make sure to use `rediss://` instead of `redis://` to use TLS for this traffic.

### OAuth Token Verification Middleware

- package location: `upload-server/internal/middleware/authverification.go`

This middleware provides OAuth 2.0 token verification for incoming requests. It currently supports JWT tokens with the plan to add support for
opaque tokens. You can use it to protect either your entire router or individual routes.

#### Configuration

You need to configure the middleware by setting up the following environment variables for your OAuth settings:

```
OAUTH_AUTH_ENABLED=true            # Enable or disable OAuth token verification
OAUTH_ISSUER_URL=https://issuer.url # URL of the token issuer
OAUTH_REQUIRED_SCOPES="scope1 scope2" # Space-separated list of required scopes
OAUTH_INTROSPECTION_URL=https://introspection.url # (for opaque tokens)
```

#### Usage - Wrapping and Protecting the Entire Router

```
func GetRouter(uploadUrl string, infoUrl string) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("/route-1", route1Handler)
	router.HandleFunc("/route-2", route2Handler)

	// Wrap the router with the OAuth middleware
	protectedRouter := OAuthTokenVerificationMiddleware(router)

	// Return the wrapped router (as http.Handler)
	return protectedRouter
}
```

#### Usage - Wrapping and Protecting an Individual Route

```
func GetRouter(uploadUrl string, infoUrl string) http.Handler {
	router := http.NewServeMux()

	// Wrap the particular route that needs to be protected
	router.HandleFunc("/public-route", publicRouteHandler)
	router.HandleFunc("/private-route", OAuthTokenVerificationHandlerFunc(privateRouteHandler))

	return router
}
```

### Configuring basic upload routing

TODO

### Configuring Authentication

TODO

## Building the source

```go
go build ./cmd/main.go -o <binary name>
```

## Unit Testing

Before running unit tests, make sure to clean the file system with the `clean.sh` script. This removes any temparary upload and report files that the tests generated.

```go
go test ./...
```

With coverage:

```go
go test -coverprofile=c.out ./...
go tool cover -html=c.out
```

### Running locally with Azurite

This method of running locally let's you simulate a connection to an Azure data store. It uses a tool called Azurite, which simulates a storage account on your local machine.

To get started, follow [these docs](https://learn.microsoft.com/en-us/azure/storage/common/storage-use-azurite?tabs=npm%2Cblob-storage) to get Azurite installed and running. Next, you'll need to set up the upload configs whithin the simulator's blob storage. You can do this with the Azure CLI, but probably the easiest way to do this is with the Azure Storage Explorer. Get this tool installed on your machine. Once installed, sign in to your Azure -SU account and connect to Azurite. Next, create a new blob container called `upload-configs`. Finally, upload the `v1` and `v2` folders within the `upload-configs` directory of this repo to this container.

## Configuration

Configuration of the `upload-server` is managed through environment variables. As a convenience these can also be set in a file
and passed as a flag :`go run ./cmd/main.go -appconf <path to conf file>` or `upload-server -appconfig <path to conf file` if you've built the binary.

By default the `upload-server` is set to run locally using the filesystem and an in memory lock mechanism, so for most local development it is sufficient rely on the defaults.

An example set of environment variables that could be exported to connect to a locally running azurite (such as the one in the docker-compose.yml file). You can find Azurite's default storage account name and key here: [https://learn.microsoft.com/en-us/azure/storage/common/storage-use-azurite?tabs=npm%2Cblob-storage#well-known-storage-account-and-key](https://learn.microsoft.com/en-us/azure/storage/common/storage-use-azurite?tabs=npm%2Cblob-storage#well-known-storage-account-and-key).

```
# ./.env

AZURE_STORAGE_ACCOUNT=devstoreaccount1
# Default test storage key can be found here https://github.com/Azure/Azurite?tab=readme-ov-file#default-storage-account
AZURE_STORAGE_KEY=<replace me with the default azurite storage key>
AZURE_ENDPOINT=http://azurite:10000/devstoreaccount1
TUS_AZURE_CONTAINER_NAME=test
DEX_MANIFEST_CONFIG_CONTAINER_NAME=config
```

could be used as follows:

```
source .env
go run ./cmd/main.go
```

or

```
go run ./cmd/main.go -appconf ./env
```

## VS Code

.vscode/launch.json

```js
{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch Package",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            // "program": "${fileDirname}"
            "program": "cmd/main.go",
            "cwd": "${workspaceFolder}",
            "args": []
        }
    ]
}
```
