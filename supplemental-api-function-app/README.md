# Suplemental API Function App

## Setup

The following is needed in order to build and deploy this function app:

- [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli)
- [Azure Functions Core Tools](https://learn.microsoft.com/en-us/azure/azure-functions/functions-run-local?tabs=v4%2Clinux%2Ccsharp%2Cportal%2Cbash#v2)
- [gradle](https://gradle.org/install/)

## Build and Deploy

To build and deploy you can use the Azure Functions Gradle plugin. You can do this with the following command:

 ```
 gradle azureFunctionsDeploy -Dsubscription=<subcription_id> -DresourceGroup=<resource_group> -DappName=<function_app_name>
 ```
 Replace the `subscription`, `resourceGroup` and `appName` parameters with the actual values.