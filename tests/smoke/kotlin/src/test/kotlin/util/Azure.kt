package util

import com.azure.identity.ClientSecretCredential
import com.azure.storage.blob.*;

class Azure {
    companion object {
        fun getBlobServiceClient(connectionString: String): BlobServiceClient {
            return BlobServiceClientBuilder()
                .connectionString(connectionString)
                .buildClient()
        }

        fun getBlobServiceClient(storageAccountName: String, azureCredentials: ClientSecretCredential): BlobServiceClient {
            return BlobServiceClientBuilder()
                .endpoint("https://$storageAccountName.blob.core.windows.net")
                .credential(azureCredentials)
                .buildClient()
        }
    }
}