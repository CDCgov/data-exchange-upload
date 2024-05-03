package model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty

@JsonIgnoreProperties(ignoreUnknown = true)
data class CopyConfig(
    @get:JsonProperty("filename_suffix") val filenameSuffix: String? = "none",
    @get:JsonProperty("targets") val targets: List<String>
)
