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

    // Specific RetryException Classes
    public class DexRetryException : Exception
    {
        public DexRetryException(){ }

        public DexRetryException(string? message) : base(message)
        {

        }

        public DexRetryException(string? message, Exception? innerException) : base(message, innerException)
        {
        }

        protected DexRetryException(SerializationInfo info, StreamingContext context) : base(info, context)
        {
        }
    }

    public class RoutingRetryException : Exception
    {
        public RoutingRetryException() { }
        public RoutingRetryException(string? message) : base(message)
        {
        }
        public RoutingRetryException(string? message, Exception? innerException) : base(message, innerException)
        {
        }
        protected RoutingRetryException(SerializationInfo info, StreamingContext context) : base(info, context)
        {
        }
    }

    public class EdavRetryException : Exception
    {
        public EdavRetryException(){ }
        public EdavRetryException(string? message) : base(message)
        {
        }
        public EdavRetryException(string? message, Exception? innerException) : base(message, innerException)
        {
        }
        protected EdavRetryException(SerializationInfo info, StreamingContext context) : base(info, context)
        {
        }
    }
}
