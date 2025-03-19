package model

import com.fasterxml.jackson.annotation.JsonProperty

data class FileInfo(
    @get:JsonProperty("size_bytes") val sizeBytes: Long,
    @get:JsonProperty("updated_at") val updatedAt: String
)