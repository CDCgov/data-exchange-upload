using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace BulkFileUploadFunctionApp.Services
{
    public interface IFeatureManagementExecutor
    {
        public IFeatureManagementExecutor ExecuteIfEnabledAsync(string flagName, Action callback);
        public IFeatureManagementExecutor ExecuteIfDisabledAsync(string flagName, Action callback);
    }
}
