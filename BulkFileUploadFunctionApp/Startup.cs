
using Microsoft.Azure.Functions.Extensions.DependencyInjection;
using Microsoft.Extensions.DependencyInjection;
using Serilog;
using Serilog.Formatting.Compact;

[assembly: FunctionsStartup(typeof(BulkFileUploadFunctionApp.Startup))]

namespace BulkFileUploadFunctionApp
{
    public class Startup : FunctionsStartup
{
    public override void Configure(IFunctionsHostBuilder builder)
    {
        Log.Logger = new LoggerConfiguration()
            .MinimumLevel.Information()
            .Enrich.FromLogContext()
            .WriteTo.Console(new RenderedCompactJsonFormatter())
            .CreateLogger();

        builder.Services.AddLogging(lb => lb.AddSerilog(dispose: true));
    }
}
}