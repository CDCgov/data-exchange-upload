using Azure.Storage.Blobs;

using Microsoft.Azure.Functions.Extensions.DependencyInjection;
using Microsoft.Extensions.DependencyInjection;
using BulkFileUploadFunctionApp.Services;

[assembly: FunctionsStartup(typeof(BulkFileUploadFunctionApp.Startup))]
namespace BulkFileUploadFunctionApp
{
    public class Startup : FunctionsStartup
    {
        public override void Configure(IFunctionsHostBuilder builder)
        {
            // Register the BlobServiceClientFactory
            builder.Services.AddSingleton<IBlobServiceClientFactory, BlobServiceClientFactoryImpl>();

            // Register the EnvironmentVariableProvider
            builder.Services.AddSingleton<IEnvironmentVariableProvider, EnvironmentVariableProviderImpl>();

            // Register the FunctionLogger
            builder.Services.AddSingleton<IFunctionLogger, FunctionLogger>();

            
        }
    }

}