package tus

import io.tus.java.client.TusClient
import io.tus.java.client.TusURLMemoryStore
import java.net.URL

class UploadClient(url: String, private val authToken: String) {
    private val client = TusClient()

    init {
        client.uploadCreationURL = URL(url)
        client.enableResuming(TusURLMemoryStore())
        initHeaders()
    }

    private fun initHeaders() {
        val headerMap = HashMap<String, String>()
        headerMap["Authorization"] = "Bearer $authToken"
        client.headers = headerMap
    }

    fun uploadFile() {
        println("UPLOADING!")
    }
}