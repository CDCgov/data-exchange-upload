using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.Serialization;
using System.Text;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionApp.Utils
{
    public class RetryException : Exception
    {
        public BlobCopyStage Stage { get; init; }
        public RetryException()
        {
        }

        public RetryException(BlobCopyStage stage, string? message) : base(message)
        {
            Stage = stage;
        }

        public RetryException(BlobCopyStage stage, string? message, Exception? innerException) : base(message, innerException)
        {
            Stage = stage;
        }
    }

    public class WriteRetryException : RetryException
    {
        public Uri srcUri { get; init; }
        public Uri destUri { get; init; }

        public WriteRetryException(BlobCopyStage stage, Uri srcUri, Uri destUri, string? message) : base(stage, message)
        {
            this.srcUri = srcUri;
            this.destUri = destUri;
        }
    }
}
