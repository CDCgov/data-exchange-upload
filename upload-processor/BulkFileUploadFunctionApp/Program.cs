using BulkFileUploadFunctionApp;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;

using BulkFileUploadFunctionApp.Services;
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

        // Registers an implementation for the IBlobServiceClientFactory interface to be resolved as a singleton.
        services.AddSingleton<IBlobServiceClientFactory, BlobServiceClientFactoryImpl>();

        // Registers an implementation for the IEnvironmentVariableProvider interface to be resolved as a singleton.
        services.AddSingleton<IEnvironmentVariableProvider, EnvironmentVariableProviderImpl>();  
        
        services.AddSingleton<IFeatureManagementExecutor, FeatureManagementExecutor>();
    })
    .Build();

host.Run();
