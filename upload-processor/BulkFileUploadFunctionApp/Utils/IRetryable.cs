using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionApp.Utils
{
    public interface IRetryable
    {
        void DoWithRetry(Action callback);
        Task DoWithRetryAsync(Func<Task> callback);
    }
}
