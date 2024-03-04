package model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty

@JsonIgnoreProperties(ignoreUnknown = true)
data class ReportContent(
    @get:JsonProperty("schema_name") val schemaName: String,
    val offset: Int?,
    val size: Int?,
    val file_destination_blob_url:String ="",
    val file_source_blob_url: String = ""
)