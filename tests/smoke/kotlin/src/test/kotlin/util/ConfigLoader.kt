package util

import com.azure.storage.blob.BlobServiceClient
import com.fasterxml.jackson.databind.node.ObjectNode
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import model.UploadConfig

class ConfigLoader {
    companion object {
        private fun preprocessConfig(jsonInput: String): String {
            val node = jacksonObjectMapper().readTree(jsonInput)
            node["metadata_config"]["fields"].forEach { field ->
                if (!field.has("compat_field_name")) {
                    (field as ObjectNode).put("compat_field_name", "default_value")
                }
            }
            return node.toString()
        }

        fun loadUploadConfig(dexBlobClient: BlobServiceClient, useCase: String, version: String): UploadConfig {
            val uploadConfigBlobClient = dexBlobClient
                .getBlobContainerClient(Constants.UPLOAD_CONFIG_CONTAINER_NAME)
                .getBlobClient("$version/${useCase}.json")

            val rawJson = uploadConfigBlobClient.downloadContent().toString()
            val preprocessedJson = preprocessConfig(rawJson)
            return jacksonObjectMapper().readValue(preprocessedJson, UploadConfig::class.java)
        }
    }
}
