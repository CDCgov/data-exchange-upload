package model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty

@JsonIgnoreProperties(ignoreUnknown = true)
data class MetadataConfig (

    @JsonProperty("version") val version: String,
    @JsonProperty("fields") val fields: List<FieldConfig>
)


