using Microsoft.Extensions.Logging;

namespace BulkFileUploadFunctionApp.Utils
{
    internal class ExceptionUtils
    {
        private readonly ILogger _logger;
        public ExceptionUtils(ILogger logger)
        {
            _logger = logger;
        }

        public void LogErrorDetails(Exception ex)
        {
            if(ex is AggregateException) {
                foreach (var inner in ex.Flatten().InnerExceptions) {
                    _logger.LogError(inner.ToString());
                }
            } else {
                _logger.LogError(ex.GetBaseException().ToString());
            }
        }
    }
}