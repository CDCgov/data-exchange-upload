using Microsoft.Extensions.Logging;


namespace BulkFileUploadFunctionApp.Utils
{
    public interface IBlobReaderFactory
    {
        IBlobReader CreateInstance(ILogger logger);
    }

    public class BlobReaderFactory : IBlobReaderFactory
    {
        public virtual IBlobReader CreateInstance(ILogger logger)
        {
            return new BlobReader(logger);
        }
    }
}