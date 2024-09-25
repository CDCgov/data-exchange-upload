# CDC DEX Upload API Load Testing Tool
This tool allows you to upload high volumes of files to an Upload API service.  The purpose of this tool is to test the
performance of the Upload API in a high load scenario, in which there are many files being uploaded in parallel of various
sizes.

## Features:
- Generate random data to stream for any given test, even synthetic HL7
- Any number of concurrent connections of different sizes
- Verbose test result reporting
- Run with or without auth

# Usage:
```
# See usage info
go run main.go -h
```

```
# Run against a local server in benchmark mode:
go run ./...
```

```
# Quickly upload a single file against a local server.  Great for smoke testing
go run ./... -load=1
```