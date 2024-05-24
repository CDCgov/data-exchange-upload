package model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties

@JsonIgnoreProperties(ignoreUnknown = true)
data class HealthResponse(
    val status: String,
    val services: List<ServiceDependencyHealth>
)