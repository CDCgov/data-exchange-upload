package model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty

@JsonIgnoreProperties(ignoreUnknown = true)
data class MetadataConfig (

    @JsonProperty("version") val version: String,
    @JsonProperty("fields") val fields: List<FieldConfig>

)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Metadata(
    @JsonProperty("upload_id") val uploadId: String?,
    @JsonProperty("dex_ingest_datetime") val dexIngestDateTime: String?
)


