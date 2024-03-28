using System;
using Azure.Storage.Blobs;
using Azure.Identity;

namespace BulkFileUploadFunctionApp.Services
{
    // Implementations for the interfaces
    public class BlobServiceClientFactoryImpl : IBlobServiceClientFactory
    {
        public BlobServiceClient CreateBlobServiceClient(string connectionString)
        {
            if (string.IsNullOrEmpty(connectionString))
            {
                // Handle the case where connectionString is null or empty.
                // You might want to throw an exception or handle it in another way.
                throw new ArgumentException("Connection string cannot be null or empty.");
            }

            // If the connectionString is valid, create and return a BlobServiceClient instance.
            var blobServiceClient = new BlobServiceClient(connectionString);
            return blobServiceClient;


        }

        public BlobServiceClient CreateBlobServiceClient(Uri serviceUri, DefaultAzureCredential credential)
        {
            if (serviceUri == null)
            {
                throw new ArgumentNullException(nameof(serviceUri));
            }
            if (credential == null)
            {
                throw new ArgumentNullException(nameof(credential));
            }
            var blobServiceClient = new BlobServiceClient(serviceUri, credential);
            return blobServiceClient;
        }
    }
}