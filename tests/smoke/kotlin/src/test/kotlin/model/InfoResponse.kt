package model

import com.fasterxml.jackson.annotation.JsonProperty
import com.fasterxml.jackson.annotation.JsonIgnoreProperties

@JsonIgnoreProperties(ignoreUnknown = true)
data class InfoResponse(
    @get:JsonProperty("manifest") val manifest: HashMap<String, String>,
    @get:JsonProperty("file_info") val fileInfo: FileInfo,
    @get:JsonProperty("upload_status") val uploadStatus: UploadStatus,
    @get:JsonProperty("deliveries") val deliveries: List<FileDelivery>?
)
