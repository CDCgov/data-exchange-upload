package model

import com.fasterxml.jackson.annotation.JsonProperty

data class FileDelivery(
    val status: String,
    val name: String,
    val location: String,
    @get:JsonProperty("delivered_at") val deliveredAt: String,
    val issues: List<HashMap<String, String>>?
)
