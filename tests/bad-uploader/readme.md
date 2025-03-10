# CDC DEX Upload API Load Testing Tool
This tool allows you to upload high volumes of files to an Upload API service.  The purpose of this tool is to test the
performance of the Upload API in a high load scenario, in which there are many files being uploaded in parallel of various
sizes.

## Features:
- Generate random data to stream for any given test, even synthetic HL7
- Any number of concurrent connections of different sizes
- Verbose test result reporting
- Run with or without auth

## Setup:
- Install Golang 1.22 or later
- Obtain SAMS SYS credentials and get your SYS user added to the DEX API activity in both SAMS Staging and Production environments

## Configuration Options:

The tool provides several command-line flags to customize the testing environment. Below is a list of key configuration options:

| Flag               | Description                                                                 |
|--------------------|-----------------------------------------------------------------------------|
| `-load`            | Specifies the number of files to upload (default: 0).                       |
| `-size`            | Defines the size of the files to upload in MB (e.g., `10`, `100`).          |
| `-parallelism`     | Specifies the number of parallel uploads (default: 1).                      |
| `-duration`        | Specifies the duration of the test (e.g., `30s`, `5m`).                     |
| `-url`             | URL of the API endpoint to test (default: local server).                    |
| `-reports-url`     | URL to send test reports to, if needed.                                     |
| `-sams-url`        | URL of the SAMS oauth token endpoint used to fetch an auth token.                                     |
| `-username`        | SAMS SYS account username.                                     |
| `-password`        | SAMS SYS account password.                                     |
| `-info-url`        | Optional used to overwrite the endpoint used to fetch upload and delivery info.  Defaults to {url}/info.                                     |
| `-v`               | Enable verbose logging.                                     |

### Default Values:

The tool uses the following default values if certain flags are not specified:

| Flag               | Default Value                        | Description                                                                     |
|--------------------|--------------------------------------|---------------------------------------------------------------------------------|
| `load`            | `0`                            | Number of concurrent uploads (adjusts based on benchmarking logic).  |
| `size`            | `5` MB                         | Size of the files to upload, in megabytes.                           |
| `parallelism`     | `runtime.NumCPU()`             | Defaults to MAXGOPROC when set to < 1                                |
| `duration`        | `0`                            | If no duration is specified, the test runs until manually stopped.   |
| `url`             | `http://localhost:8080/files`  | Default API endpoint for file uploads.                               |
| `reports-url`     | Not set                        | No reports are sent unless explicitly specified.                     |


To view the full list of options at any time, you can run:
```bash
go run ./... -h
```

## Usage Examples:

### **Basic Upload Test With Delivery Check (Default)**
Run a basic test to verify the tool's setup with minimal configuration or for a very basic test of the upload server.

```bash
go run ./... -load=1
```

This command uploads a single 5MB file to the local (default) upload server endpoint.

### **Providing an Upload Endpoint URL**
Run a basic test to verify the tool's setup with minimal configuration or for a very basic test of the upload server at a specified endpoint.

```bash
go run ./... -load=1 -url=https://upload-api.server:8080/files
```

This command uploads a single 5MB file to the specified upload server endpoint.

### **Parallel Custom Size Uploads Test**
Test the system's performance with concurrent file uploads of a custom size.

```bash
go run ./... -load=50 -parallelism=8 -size=20 -url=https://upload-api.server:8080/files
```

This example uploads 50 files, each 20MB in size, with 8 uploads occurring in parallel.

### **Test with Report URL**
Run a test and send the results to a reporting endpoint.

```bash
go run ./... -load=10 -size=50 -url=https://upload-api.server:8080/files -reports-url=https://reports-server:8080/reports
```

This command uploads 10 files, each 50MB in size, and sends the test results to the report server.

## Smoke Testing:
The following commands are useful for performing smoke tests against deployed environments.  The following URLs can be used to test both from public internet and internal to CDC network:

| URL               | Environment| Public/Internal |
|-------------------|------------|-----------------|
| https://apidev.cdc.gov/upload | dev | Public |
| https://uploaddev.ocio-eks-dev-ede.cdc.gov/files | dev | Internal |
| https://apitst.cdc.gov/upload | tst | Public |
| https://upload.phdo-eks-test.cdc.gov/files | tst | Internal |
| https://apistg.cdc.gov/upload | stg | Public |
| https://upload.phdo-eks-stg.cdc.gov/files | stg | Internal |
| https://api.cdc.gov/upload | prd | Public |
| https://upload.phdo-eks-prd.cdc.gov/files | prd | Internal |

### Minimal Command for Public URL

```bash
go run ./... -load=1 -url=https://api.cdc.gov/upload -info-url=https://api.cdc.gov/upload/info -sams-url=https://api.cdc.gov/oauth -username=*** -password=***
```

**Note that the `info-url` flag is required when testing against the public URL due to the `/upload` path prefix.**

### Minimal Command for Internal URL (AWS EKS Ingress)

```bash
go run ./... -load=1 -url=https://upload.phdo-eks-prd.cdc.gov/files -sams-url=https://api.cdc.gov/oauth -username=*** -password=***
```

**Note that SAMS credentials are required for testing gainst the internal URL since JWT authentication is now enabled in all environments.**
**Note that these commands can only be run from within the CDC network.**
