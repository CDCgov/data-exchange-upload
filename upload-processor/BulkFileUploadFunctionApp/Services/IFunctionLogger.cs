
namespace BulkFileUploadFunctionApp.Services
{
    public interface IFunctionLogger<T>
    {
        void LogInformation(string message);
        void LogError(string message);
        void LogError(Exception ex, string message);
    }
}