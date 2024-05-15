package tus

import dex.DexTusClient
import io.tus.java.client.*
import java.io.File
import java.io.IOException
import java.net.URL

class UploadClient(url: String, private val authToken: String) {
    private val client = DexTusClient()

    init {
        client.uploadCreationURL = URL("$url/upload")
        client.enableResuming(TusURLMemoryStore())
        initHeaders()
        System.setProperty("sun.net.http.allowRestrictedHeaders", "true")
    }

    private fun initHeaders() {
        val headerMap = HashMap<String, String>()
        headerMap["Authorization"] = "Bearer $authToken"
        headerMap["Content-Length"] = "0"
        client.headers = headerMap
    }

    fun uploadFile(file: File, metadata: Map<String, String>, chunkSize: Int = 1024): String? {
        var uploadId: String? = null
        val uploadHandle = TusUpload(file).apply {
            setMetadata(metadata)
        }

        val executor = object : TusExecutor() {
            @Throws(ProtocolException::class, IOException::class)
            override fun makeAttempt() {
                try {
                    val uploader = client.resumeOrCreateUpload(uploadHandle)
                    uploader.chunkSize = chunkSize

                    do {
                        val totalBytes = uploadHandle.size
                        val bytesUploaded = uploader.offset
                        val progress = bytesUploaded.toDouble() / totalBytes * 100
                        println(String.format("Upload at %06.2f%%.", progress))
                    } while (uploader.uploadChunk() > -1)

                    uploader.finish()
                    uploadId = parseUploadIdFromUrl(uploader.uploadURL.toString())
                    println("Upload finished.")
                    println(String.format("Upload available at: %s", uploader.uploadURL.toString()))
                } catch (e: ProtocolException) {
                    var errorMessage = e.message
                    client.connection.errorStream.let {
                        errorMessage = "$errorMessage ; response: ${it.readAllBytes().toString(Charsets.UTF_8)}"
                    }

                    throw ProtocolException(errorMessage, client.connection)
                }
            }
        }

        executor.makeAttempts()

        return uploadId
    }

    private fun parseUploadIdFromUrl(uploadUrl: String): String {
        return uploadUrl.split("/").last().trim()
    }
}