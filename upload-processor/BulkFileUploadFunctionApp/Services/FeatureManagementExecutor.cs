using Microsoft.Extensions.Configuration.AzureAppConfiguration;
using Microsoft.FeatureManagement;
using System;
using Microsoft.Extensions.Configuration;

namespace BulkFileUploadFunctionApp.Services
{
    public class FeatureManagementExecutor : IFeatureManagementExecutor
    {
        private readonly IConfigurationRefresher _configurationRefresher;
        private readonly IConfiguration _configuration;

        public FeatureManagementExecutor(IConfigurationRefresherProvider configurationRefresherProvider, IConfiguration configuration) 
        {
            _configurationRefresher = configurationRefresherProvider.Refreshers.First();
            _configuration = configuration;
        }
        public IFeatureManagementExecutor ExecuteIfEnabledAsync(string flagName, Action callback)
        {
            _configurationRefresher.TryRefreshAsync();
            bool isEnabled = _configuration.GetValue<bool>($"FeatureManagement:{flagName}");

            if (isEnabled)
            {
                callback();
            }

            return this;
        }

        public IFeatureManagementExecutor ExecuteIfDisabledAsync(string flagName, Action callback)
        {
            _configurationRefresher.TryRefreshAsync();
            bool isEnabled = _configuration.GetValue<bool>($"FeatureManagement:{flagName}");

            if (!isEnabled)
            {
                callback();
            }

            return this;
        }
    }
}
