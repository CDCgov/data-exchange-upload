package model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty

@JsonIgnoreProperties(ignoreUnknown = true)
data class ReportContent(
    @get:JsonProperty("schema_name") val schemaName: String,
    @get:JsonProperty("destination") val destination: String?,
    val offset: Int?,
    val size: Int?
)