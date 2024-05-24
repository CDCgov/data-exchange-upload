package model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty

@JsonIgnoreProperties(ignoreUnknown = true)
data class ServiceDependencyHealth(
    val service: String,
    val status: String,
    @get:JsonProperty("health_issue") val healthIssue: String
)
