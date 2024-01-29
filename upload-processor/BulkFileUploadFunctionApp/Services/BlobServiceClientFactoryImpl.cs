using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using Azure.Storage.Blobs;

namespace BulkFileUploadFunctionApp
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
            throw new ArgumentException("Connection string cannot be null or empty.", nameof(connectionString));
        }

        // If the connectionString is valid, create and return a BlobServiceClient instance.
        var blobServiceClient = new BlobServiceClient(connectionString);
        return blobServiceClient;

        
    }
}
}