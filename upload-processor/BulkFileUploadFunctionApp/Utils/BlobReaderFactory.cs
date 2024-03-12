using Microsoft.Extensions.Logging;


namespace BulkFileUploadFunctionApp.Utils
{
    public class BlobReaderFactory
    {
        public virtual IBlobReader CreateInstance(ILogger logger)
        {
            return new BlobReader(logger);
        }
    }
}