using BulkFileUploadFunctionApp;
using BulkFileUploadFunctionApp.Service;
using Microsoft.Azure.Functions.Worker;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;

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
            builder.AddAzureAppConfiguration(cs);
        })
    .ConfigureFunctionsWorkerDefaults()
    .ConfigureServices(services =>
    {
        services.AddApplicationInsightsTelemetryWorkerService();
        services.ConfigureFunctionsApplicationInsights();
        services.AddHttpClient<IProcStatService, ProcStatusService>(client =>
        {
          client.BaseAddress = new Uri(Environment.GetEnvironmentVariable("PROC_STATUS_URL") ?? "");
        });
        services.AddSingleton<IProcStatService, ProcStatusService>();
    })
    .Build();

host.Run();
