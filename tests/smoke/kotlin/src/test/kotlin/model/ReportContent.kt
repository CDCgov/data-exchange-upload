package model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty

@JsonIgnoreProperties(ignoreUnknown = true)
data class ReportContent(
    @get:JsonProperty("content_schema_name") val contentSchemaName: String,
    @get:JsonProperty("content_schema_version") val contentSchemaVersion: String,
    val transforms: List<Transform>?,
    val metadata: Metadata?,
    @get:JsonProperty("file_source_blob_url") val fileSourceBlobUrl: String?,
    @get:JsonProperty("file_destination_blob_url") val fileDestinationBlobUrl: String?,
    @get:JsonProperty("destination_name") val destinationName: String?,
    val status: String?,
    val filename: String?,
    val tguid: String?,
    val offset: Int?,
    val size: Int?
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Transform(
    val action: String?,
    val field: String?,
    val value: String?
)


