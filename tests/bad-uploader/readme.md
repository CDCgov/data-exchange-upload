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
# see usage info
go run main.go -h
```

```
# run against a local server in benchmark mode:
go run ./...
```