using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp.Utils
{
    internal class ExceptionUtils
    {
        public static void LogErrorDetails(Exception ex, ILogger _logger)
        {
            if(ex is AggregateException) {
                AggregateException ae = (AggregateException)ex;
                foreach (Exception innerException in ae.Flatten().InnerExceptions) {
                    _logger.LogError(innerException.ToString());
                }
            } else {
                _logger.LogError(ex.GetBaseException().ToString());
            }
        }
    }
}