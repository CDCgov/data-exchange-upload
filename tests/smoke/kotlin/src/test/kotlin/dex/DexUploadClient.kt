package dex

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.KotlinModule
import com.fasterxml.jackson.module.kotlin.readValue
import model.AuthResponse
import model.HealthResponse
import model.InfoResponse
import okhttp3.FormBody
import okhttp3.OkHttpClient
import okhttp3.Request
import java.io.IOException
import java.nio.charset.StandardCharsets
import util.EnvConfig
import org.slf4j.LoggerFactory

class DexUploadClient(private val url: String) {
    private val httpClient = OkHttpClient()
    private val objectMapper = ObjectMapper().registerModule(KotlinModule.Builder().build())
    private val logger = LoggerFactory.getLogger(DexUploadClient::class.java)

    fun getToken(username: String, password: String): String {
        val body = FormBody.Builder(StandardCharsets.UTF_8)
            .add("username", username)
            .add("password", password)
            .build()

        val req = Request.Builder()
            .url("$url/oauth")
            .post(body)
            .build()

        httpClient.newCall(req).execute().use { response ->
            if (!response.isSuccessful) {
                throw IOException("Error getting token: ${response.message}")
            }

            val responseBody = response.body?.string()
                ?: throw IOException("Empty response body from server")

            val authResponse: AuthResponse = objectMapper.readValue(responseBody)
            return authResponse.accessToken
        }
    }

    fun getFileInfo(id: String, authToken: String): InfoResponse {

        val req = Request.Builder()
            .url("${EnvConfig.INFO_URL}/info/$id")
            .header("Authorization", "Bearer $authToken")
            .build()

        Thread.sleep(1000) // wait to load delivery response

        httpClient.newCall(req).execute().use { response ->
            if (!response.isSuccessful) {
                throw IOException("Error getting file info. ${response.message}")
            }

            val responseBody = response.body?.string() ?: throw IOException("Empty response body from server")

            logger.info("Raw response body: $responseBody")

            return objectMapper.readValue(responseBody, InfoResponse::class.java)
        }
    }

    fun getHealth(authToken: String): HealthResponse {
        val req = Request.Builder()
            .url("$url/upload/health")
            .header("Authorization", "Bearer $authToken")
            .build()

        httpClient.newCall(req).execute().use { response ->
            if (!response.isSuccessful) {
                throw IOException("Error getting health check: ${response.message}")
            }

            val responseBody = response.body?.string()
                ?: throw IOException("Empty response body from server")

            return objectMapper.readValue(responseBody)
        }
    }
}