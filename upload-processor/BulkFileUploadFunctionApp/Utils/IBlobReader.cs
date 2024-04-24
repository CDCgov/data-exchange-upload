namespace BulkFileUploadFunctionApp.Utils
{
    public interface IBlobReader
    {
        Task<T?> GetObjectFromBlobJsonContent<T>(string connectionString, string sourceContainerName, string blobPathname);
    }
}