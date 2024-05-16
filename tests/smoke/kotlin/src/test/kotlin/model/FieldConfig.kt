package model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty

@JsonIgnoreProperties(ignoreUnknown = true)
data class FieldConfig(
    @JsonProperty("field_name") val fieldName: String,
    @JsonProperty("compat_field_name") val compatFieldName: String? = null
)