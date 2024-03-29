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
        public RetryException()
        {
        }

        public RetryException(string? message) : base(message)
        {

        }

        public RetryException(string? message, Exception? innerException) : base(message, innerException)
        {
        }

        protected RetryException(SerializationInfo info, StreamingContext context) : base(info, context)
        {
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
