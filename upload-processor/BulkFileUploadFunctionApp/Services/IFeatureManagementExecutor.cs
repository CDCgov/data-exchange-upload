namespace BulkFileUploadFunctionApp.Services
{
    public interface IFeatureManagementExecutor
    {
        public Task<IFeatureManagementExecutor> ExecuteIfEnabledAsync(string flagName, Func<Task> callback);
        public IFeatureManagementExecutor ExecuteIfEnabled(string flagName, Action callback);
        public Task<IFeatureManagementExecutor> ExecuteIfDisabledAsync(string flagName, Func<Task> callback);
        public IFeatureManagementExecutor ExecuteIfDisabled(string flagName, Action callback);
    }
}