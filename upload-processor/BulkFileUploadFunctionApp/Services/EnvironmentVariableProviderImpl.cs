

namespace BulkFileUploadFunctionApp.Services
{
    public class EnvironmentVariableProviderImpl : IEnvironmentVariableProvider
    {
        public string GetEnvironmentVariable(string name)
        {
            if (string.IsNullOrEmpty(name))
            {
                // Handle the case where the name is null or empty.
                // You might want to return a default value or handle it in another way.
                throw new ArgumentException("Environment variable name cannot be null or empty.", nameof(name));
            }

            // Retrieve the environment variable. 
            // This will return null if the environment variable does not exist.
            string value = Environment.GetEnvironmentVariable(name);

            // Optionally, you might want to handle the case where the environment variable is not found.
            // For example, you can return a default value, log a warning, or throw an exception.
            if (string.IsNullOrEmpty(value))
            {
                // Handle the case where the environment variable does not exist       
                throw new KeyNotFoundException($"Environment variable '{name}' not found.");
            }

            // Return the retrieved value.
            return value;
        }
    }
}