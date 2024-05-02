package model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty

@JsonIgnoreProperties(ignoreUnknown = true)
data class UploadConfig(
    @get:JsonProperty("copy_config") val copyConfig: CopyConfig,
)
