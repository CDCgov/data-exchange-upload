using BulkFileUploadFunctionApp;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using BulkFileUploadFunctionApp.Services;
using BulkFileUploadFunctionApp.Utils;
using Microsoft.FeatureManagement;

var host = new HostBuilder()
    .ConfigureLogging(builder =>
    {
        var config = new JsonLoggerConfiguration
        {
            LogLevel = LogLevel.Information,
            TimestampFormat = "yyyy-MM-dd HH:mm:ss.fff"
            // Configure other properties as needed
        };
        builder.ClearProviders();
        builder.AddProvider(new JsonLoggerProvider(config));
    })
    .ConfigureAppConfiguration(builder =>
        {
            string cs = Environment.GetEnvironmentVariable("FEATURE_MANAGER_CONNECTION_STRING") ?? "";
            builder.AddAzureAppConfiguration(options =>
            {
                options.Connect(cs)
                       .UseFeatureFlags();
            });
        })
    .ConfigureFunctionsWorkerDefaults()
    .ConfigureServices(services => {
        services.AddApplicationInsightsTelemetryWorkerService();
        services.ConfigureFunctionsApplicationInsights();
        services.AddAzureAppConfiguration();
        services.AddFeatureManagement();

        // Register Proc Stat Http Service.
        services.AddHttpClient<IProcStatClient, ProcStatClient>(client =>
        {
            client.BaseAddress = new Uri(Environment.GetEnvironmentVariable("PS_API_URL") ?? "");
        });

        services.AddSingleton<IUploadEventHubService, UploadEventHubService>();

        services.AddSingleton<IUploadProcessingService, UploadProcessingService>();

        // Registers an implementation for the IBlobServiceClientFactory interface to be resolved as a singleton.
        services.AddSingleton<IBlobClientFactory, BlobClientFactory>();

        // Registers an implementation for the IBlobManagementService interface to be resolved as a singleton.
        services.AddSingleton<IBlobManagementService, BlobManagementService>();

        // Registers an implementation for the IEnvironmentVariableProvider interface to be resolved as a singleton.
        services.AddSingleton<IEnvironmentVariableProvider, EnvironmentVariableProviderImpl>();       

        // Registers an implementation for the IBlobCopyHelper interface to be resolved as a singleton.
        services.AddSingleton<IBlobCopyHelper, BlobCopyHelper>();

        // Registers an implementation for the BlobReaderFactory's IBlobReader interface to be resolved as a singleton.
        services.AddSingleton<IBlobReaderFactory, BlobReaderFactory>();
        
        services.AddSingleton<IFeatureManagementExecutor, FeatureManagementExecutor>();
    })
    .Build();

host.Run();
