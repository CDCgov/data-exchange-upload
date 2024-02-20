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

        public IFeatureManagementExecutor ExecuteIfEnabled(string flagName, Action callback)
        {
            _configurationRefresher.TryRefreshAsync();
            bool isEnabled = _configuration.GetValue<bool>($"FeatureManagement:{flagName}");


            if (isEnabled)
            {
                callback();
            }

            return this;
        }

        public async Task<IFeatureManagementExecutor> ExecuteIfEnabledAsync(string flagName, Func<Task> callback)
        {
            await _configurationRefresher.TryRefreshAsync();
            bool isEnabled = _configuration.GetValue<bool>($"FeatureManagement:{flagName}");


            if (isEnabled)
            {
                await callback();
            }

            return this;
        }

        public async Task<IFeatureManagementExecutor> ExecuteIfDisabledAsync(string flagName, Func<Task> callback)
        {
            await _configurationRefresher.TryRefreshAsync();
            bool isEnabled = _configuration.GetValue<bool>($"FeatureManagement:{flagName}");

            if (!isEnabled)
            {
                await callback();
            }

            return this;
        }

        public IFeatureManagementExecutor ExecuteIfDisabled(string flagName, Action callback)
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
