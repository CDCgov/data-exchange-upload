package util

import com.azure.storage.blob.BlobServiceClient
import com.azure.storage.blob.models.BlobStorageException
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import model.UploadConfig

class ConfigLoader {
    companion object {
        fun loadUploadConfig(dexBlobClient: BlobServiceClient, filename: String, version: String): UploadConfig {
            try {
                val uploadConfigBlobClient = dexBlobClient
                    .getBlobContainerClient(Constants.UPLOAD_CONFIG_CONTAINER_NAME)
                    .getBlobClient("$version/$filename")

                val rawJson = uploadConfigBlobClient.downloadContent().toString()
                return jacksonObjectMapper().readValue(rawJson, UploadConfig::class.java)
            }
            catch (e: BlobStorageException) {
                throw BlobStorageException("${e.message} Could not find $version/$filename", e.response, e.value)
            }
        }
    }
}
