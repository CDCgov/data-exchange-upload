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

        fun loadUploadConfig(dexBlobClient: BlobServiceClient, manifest: HashMap<String, String>): UploadConfig {
            val versionFolder = when (manifest["version"]) {
                "2.0" -> "v2"
                else -> "v1"
            }
            val filename = when (manifest["version"]) {
                "2.0" -> "${manifest["data_stream_id"]}-${manifest["data_stream_route"]}.json"
                else -> "${manifest["meta_destination_id"]}-${manifest["meta_ext_event"]}.json"
            }

            try {
                val uploadConfigBlobClient = dexBlobClient
                    .getBlobContainerClient(Constants.UPLOAD_CONFIG_CONTAINER_NAME)
                    .getBlobClient("$versionFolder/$filename")

                val rawJson = uploadConfigBlobClient.downloadContent().toString()
                return jacksonObjectMapper().readValue(rawJson, UploadConfig::class.java)
            } catch (e: BlobStorageException) {
                throw BlobStorageException("${e.message} Could not find $versionFolder/$filename", e.response, e.value)
            }
        }
    }
}
