# Azure Blob Storage Script

This Go script connects to an Azure Storage Account, lists blobs within a specified container, and deletes them.

## Prerequisites

- Azure Storage Account credentials:
  - `TARGET_ENV`: Target Environment
  - `ACCOUNT_KEY`: The access key for the Storage Account (set as an environment variable)
  - `DATA_STREAM`: The data stream name
  - `ROUTE`: The route name

## How To run

go run main.go -target_env="target-env" -data_stream="data-stream-name" -route="route-name"