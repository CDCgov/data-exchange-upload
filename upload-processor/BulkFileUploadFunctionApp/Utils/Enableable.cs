using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using BulkFileUploadFunctionApp.Services;

namespace BulkFileUploadFunctionApp.Utils
{
    public abstract class Enableable
    {
        protected string? FeatureFlagKey { get; init; }
        protected IFeatureManagementExecutor? Executor { get; init; }
        public abstract void DoIfEnabled(Action callback);
        public abstract Task DoIfEnabledAsync(Func<Task> callback);
    }
}
