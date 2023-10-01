namespace BulkFileUploadFunctionApp.Utils
{
    internal class ExceptionUtils
    {
        public LogErrorDetails(Exception ex)
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