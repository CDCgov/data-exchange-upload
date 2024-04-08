Run the following to enable these integration tests against a local azurite instance provided by the docker-compose.yml file.

```
podman-compose up -d
UPLOAD_INTEGRATION_TEST=true go test ./...
```