# DEX TUSD Go Server 
A resumable file upload server for OCIO Data Exchange (DEX)

## Folder structure
Repo is structured (as feasible) based on the [golang-standards/project-layout](https://github.com/golang-standards/project-layout)

## References
- Based on the [tus](https://tus.io/) open protocol for resumable file uploads
- Based on the [tusd](https://github.com/tus/tusd) official reference implementation

## Usage

### Configuring the storage backend
This service currently supports local file system, Azure, and AWS as storage backends.  Configuring a storage backend
for your setup is done via environment variables.  Environment variables can be set at the system level, or via a `.env` file
located within the `configs/local/` directory.  Here are some examples for configuring the different storage backends
that this service supports.

### Local file system
By default, this service uses the file system of the host machine it is running on as a storage backend.  Therefore, no
environment variables are necessary to set.  To run, simply execute
```go
go run ./cmd/main.go
```
This will start the HTTP server at http://localhost:8080.  With a Tus client, you can upload files to http://localhost:8080/files,
and they will show up in the `uploads/` directory.

You can configure this behavior with the following environment variables:
- `SERVER_PORT` - Sets the port that the server runs on.  Default is 8080.
- `LOCAL_FOLDER_UPLOADS_TUS` - Relative path to the folder where tus will upload files to.  Default is `./uploads`.
- `TUSD_HANDLER_BASE_PATH` - URL path for the file upload endpoint.  Default is `/files` or `/files/`.
- `TUS_UPLOAD_PREFIX` - Sub folder to drop files into within the local or cloud folder.  Defaults to `tus-prefix`.



## Building the source
```go
go build ./cmd/main.go -o <binary name>
```

## Unit Testing
Before running unit tests, make sure to clean the file system with the `clean.sh` script.  This removes any temparary upload and report files that the tests generated.

```go
go test ./...
```

With coverage:
```go
go test -coverprofile=c.out ./...
go tool cover -html=c.out
```

### Running locally with Azurite

This method of running locally let's you simulate a connection to an Azure data store.  It uses a tool called Azurite, which simulates a storage account on your local machine.

To get started, follow [these docs](https://learn.microsoft.com/en-us/azure/storage/common/storage-use-azurite?tabs=npm%2Cblob-storage) to get Azurite installed and running.  Next, you'll need to set up the upload configs whithin the simulator's blob storage.  You can do this with the Azure CLI, but probably the easiest way to do this is with the Azure Storage Explorer.  Get this tool installed on your machine.  Once installed, sign in to your Azure -SU account and connect to Azurite.  Next, create a new blob container called `upload-configs`.  Finally, upload the `v1` and `v2` folders within the `upload-configs` directory of this repo to this container.

## Configuration

Configuration of the `upload-server` is managed through environment variables. As a convenience these can also be set in a file
and passed as a flag :`go run ./cmd/main.go -appconf <path to conf file>` or `upload-server -appconfig <path to conf file` if you've built the binary.

By default the `upload-server` is set to run locally using the filesystem and an in memory lock mechanism, so for most local development it is sufficient rely on the defaults.

An example set of environment variables that could be exported to connect to a locally running azurite (such as the one in the docker-compose.yml file).  You can find Azurite's default storage account name and key here: [https://learn.microsoft.com/en-us/azure/storage/common/storage-use-azurite?tabs=npm%2Cblob-storage#well-known-storage-account-and-key](https://learn.microsoft.com/en-us/azure/storage/common/storage-use-azurite?tabs=npm%2Cblob-storage#well-known-storage-account-and-key).

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

