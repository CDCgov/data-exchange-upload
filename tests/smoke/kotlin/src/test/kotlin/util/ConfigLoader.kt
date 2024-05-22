package util

import com.azure.storage.blob.BlobServiceClient
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import model.UploadConfig

class ConfigLoader {
    companion object {
        fun loadUploadConfig(dexBlobClient: BlobServiceClient, filename: String, version: String): UploadConfig {
            val uploadConfigBlobClient = dexBlobClient
                .getBlobContainerClient(Constants.UPLOAD_CONFIG_CONTAINER_NAME)
                .getBlobClient("$version/$filename")

            val rawJson = uploadConfigBlobClient.downloadContent().toString()
            return jacksonObjectMapper().readValue(rawJson, UploadConfig::class.java)
        }
    }
}
