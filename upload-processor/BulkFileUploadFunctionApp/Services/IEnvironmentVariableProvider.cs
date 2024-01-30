

namespace BulkFileUploadFunctionApp.Services
{
    public interface IEnvironmentVariableProvider
    {
        string GetEnvironmentVariable(string name);
    }
}