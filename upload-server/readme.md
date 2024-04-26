# DEX TUSD Go Server 
A resumable file upload server for OCIO Data Exchange (DEX)

## Folder structure
Repo is structured (as feasible) based on the [golang-standards/project-layout](https://github.com/golang-standards/project-layout)

## References
- Based on the [tus](https://tus.io/) open protocol for resumable file uploads
- Based on the [tusd](https://github.com/tus/tusd) official reference implementation

## Running and Building

### Running locally
```go
go run ./cmd/main.go
```

### Building
```go
go build ./cmd/main.go -o <binary name>
```

### Unit Testing
```go
go test ./...
```

With coverage:
```go
go test -coverprofile=c.out ./...
go tool cover -html=c.out
```

## Configuration

Configuration of the `upload-server` is managed through environment variables. As a convenience these can also be set in a file
and passed as a flag :`go run ./cmd/main.go -appconf <path to conf file>` or `upload-server -appconfig <path to conf file` if you've built the binary.

By default the `upload-server` is set to run locally using the filesystem and an in memory lock mechanism, so for most local development it is sufficient rely on the defaults.

An example set of environment variables that could be exported to connect to a locally running azurite (such as the one in the docker-compose.yml file)

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

