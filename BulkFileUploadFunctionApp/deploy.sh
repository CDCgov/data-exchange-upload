ENV=$1

func azure functionapp publish $ENV-upload-processor --csharp --force
az functionapp config appsettings set --name $ENV-upload-processor --resource-group ocio-ede-$ENV-moderate-hl7-rg --settings "FUNCTIONS_WORKER_RUNTIME=dotnet-isolated"
