package model

import com.fasterxml.jackson.annotation.JsonProperty

data class UploadStatus(
    val status: String,
    @get:JsonProperty("chunk_received_at") val chunkReceivedAt: String
)