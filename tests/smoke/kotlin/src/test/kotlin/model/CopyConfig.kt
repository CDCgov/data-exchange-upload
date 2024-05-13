package model

import com.azure.storage.blob.BlobServiceClient
import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import util.Azure
import util.Constants
import util.EnvConfig

@JsonIgnoreProperties(ignoreUnknown = true)
data class CopyConfig(
    @get:JsonProperty("filename_suffix") val filenameSuffix: String? = "none",
    @get:JsonProperty("targets") val targets: List<String>
)

