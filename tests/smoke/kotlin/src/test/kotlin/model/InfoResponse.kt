package model

import com.fasterxml.jackson.annotation.JsonProperty

data class InfoResponse(
    val manifest: HashMap<String, String>,
    @get:JsonProperty("file_info") val fileInfo: FileInfo
)
