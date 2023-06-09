using System.Runtime.Serialization;

namespace BulkFileUploadFunctionApp
{
    [Serializable]
    internal class UploadConfigException : Exception
    {
        public UploadConfigException()
        {
        }

        public UploadConfigException(string? message) : base(message)
        {
        }

        public UploadConfigException(string? message, Exception? innerException) : base(message, innerException)
        {
        }

        protected UploadConfigException(SerializationInfo info, StreamingContext context) : base(info, context)
        {
        }
    }
}
