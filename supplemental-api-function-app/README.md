# Suplemental API Function App

## Setup

The following is needed in order to build and deploy this function app:

- [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli)
- [Azure Functions Core Tools](https://learn.microsoft.com/en-us/azure/azure-functions/functions-run-local?tabs=v4%2Clinux%2Ccsharp%2Cportal%2Cbash#v2)
- [gradle](https://gradle.org/install/)

## Run Locally

You can run this function app locally using the Gradle Azure plugin, which uses Azure function core tools at a lower level.
You'll also need to log into the Azure CLI so the function app assumes your service principal when it runs.  After you've done that,
you can run locally with the following steps:

1. Create a `local.settings.json` file.  This is so you can provide runtime environment variables.  Here is an example for the
dev environment.
```json
{
  "Values": {
    "DexStorageEndpoint": "https://ocioededataexchangedev.blob.core.windows.net",
    "DexStorageConnectionString": "***",
    "TusHooksContainerFileName": "tusd-file-hooks",
    "DestinationsFileName": "allowed_destinations_and_events.json"
  }
}
```
**Note that the connection string has been redacted.  You will need to get it from the storage account in Azure.**
2. Build with `gradle jar`
3. Run with `gradle azureFunctionsRun -DresourceGroup=... -DappName=... -Dsubscription=...`

**Note that the resourceGroup, appName, and subscription are required environment variables, and can be found on the remote
function app's overview page**

## Build and Deploy

To build and deploy you can use the Azure Functions Gradle plugin. You can do this with the following command:

 ```
 gradle azureFunctionsDeploy -Dsubscription=<subcription_id> -DresourceGroup=<resource_group> -DappName=<function_app_name>
 ```
 Replace the `subscription`, `resourceGroup` and `appName` parameters with the actual values.