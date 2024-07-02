package org.example.model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty

@JsonIgnoreProperties(ignoreUnknown = true)
data class Info(
    @get:JsonProperty("MetaData") val metadata: HashMap<String, String>
)
