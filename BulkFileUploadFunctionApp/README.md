# Upload Function App

This function app is triggered on a new file upload through tus, and is responsible for validating, applying approprate metadata, and copying that file to the appropriate blob container in EDAV.

## Setup

The following is needed in order to build and deploy this function app:

- [.NET 6.0 SDK](https://dotnet.microsoft.com/en-us/download)
- [Azure Functions Core Tools](https://learn.microsoft.com/en-us/azure/azure-functions/functions-run-local?tabs=v4%2Clinux%2Ccsharp%2Cportal%2Cbash#v2)
- [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli)

## Compile and Publish the Project

### Azure CLI

You can compile and publish this project with the following Azure CLI command:

```
func azure functionapp publish ocio-ede-<env>-bulk-file-upload-processor --csharp --force
```

Where `<env>` is the name of the environment you are deploying to. Options are `dev` or `prd`.

Finally, set the `FUNCTION_WORKER_RUNTIME` setting to `dotnet-isolated`.

```
az functionapp config appsettings set --name ocio-ede-dev-bulk-file-upload-processor --resource-group ocio-ede-dev-moderate-hl7-rg --settings "FUNCTIONS_WORKER_RUNTIME=dotnet-isolated"
```
