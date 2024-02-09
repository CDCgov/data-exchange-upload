package model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty

@JsonIgnoreProperties(ignoreUnknown = true)
data class Report (
    @get:JsonProperty("stage_name") val stageName: String,
    val issues: Array<String>?
)