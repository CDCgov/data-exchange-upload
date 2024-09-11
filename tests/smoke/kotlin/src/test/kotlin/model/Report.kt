package model

import com.fasterxml.jackson.annotation.JsonIgnoreProperties
import com.fasterxml.jackson.annotation.JsonProperty

@JsonIgnoreProperties(ignoreUnknown = true)
data class DataResponse(
    @JsonProperty("data") val data: ReportResponse
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class ReportResponse(
    @JsonProperty("getReports") val reports: List<Report>
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class Report(
    val content: ReportContent,
    val contentType: String?,
    val data: String?,
    val dataStreamId: String?,
    val dataStreamRoute: String?,
    val dexIngestDateTime: String?,
    val id: String?,
    val jurisdiction: String?,
    val reportId: String?,
    val senderId: String?,
    val tags: List<String>?,
    val timestamp: String?,
    val uploadId: String?,
    @JsonProperty("stageInfo") val stageInfo: StageInfo?
)

@JsonIgnoreProperties(ignoreUnknown = true)
data class StageInfo(
    @JsonProperty("action") val action: String?,
    @JsonProperty("status") val status: String?,
    @JsonProperty("issues") val issues: List<String>?,
    @JsonProperty("startProcessingTime") val startProcessingTime: String?,
    @JsonProperty("endProcessingTime") val endProcessingTime: String?,
    @JsonProperty("service") val service: String?,
    @JsonProperty("version") val version: String?
)


