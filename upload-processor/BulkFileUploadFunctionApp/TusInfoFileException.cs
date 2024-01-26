// Default URL for triggering event grid function in the local environment.
// http://localhost:7071/runtime/webhooks/EventGrid?functionName={functionname}
using System.Runtime.Serialization;

namespace BulkFileUploadFunctionApp
{
    [Serializable]
    internal class TusInfoFileException : Exception
    {
        public TusInfoFileException()
        {
        }

        public TusInfoFileException(string? message) : base(message)
        {
        }

        public TusInfoFileException(string? message, Exception? innerException) : base(message, innerException)
        {
        }

        protected TusInfoFileException(SerializationInfo info, StreamingContext context) : base(info, context)
        {
        }
    }
}