package dex

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.KotlinModule
import model.AuthResponse
import model.InfoResponse
import okhttp3.FormBody
import okhttp3.Headers
import okhttp3.OkHttpClient
import okhttp3.Request
import okio.IOException
import java.nio.charset.Charset
import java.nio.charset.StandardCharsets

class DexUploadClient(private val url: String) {
    private val httpClient = OkHttpClient()
    private val objectMapper = ObjectMapper().registerModule(KotlinModule.Builder().build())

    fun getToken(username: String, password: String): String {
        val body = FormBody.Builder(Charset.forName(StandardCharsets.UTF_8.name()))
            .add("username", username)
            .add("password", password)
            .build()

        val req = Request.Builder()
            .url("$url/oauth")
            .post(body)
            .build()

        val resp = httpClient
            .newCall(req)
            .execute()

        if (!resp.isSuccessful) {
            throw IOException("Error getting token.")
        }

        val respBody: AuthResponse = objectMapper.readValue(resp.body?.string()
            ?: throw IOException("Empty SAMS response"), AuthResponse::class.java)

        return respBody.accessToken
    }

    fun getFileInfo(id: String, authToken: String): InfoResponse {
        val req = Request.Builder()
            .url("$url/upload/info/$id")
            .header("Authorization", "Bearer $authToken")
            .build()

        val resp = httpClient
            .newCall(req)
            .execute()

        if (!resp.isSuccessful) {
            throw IOException("Error getting file info. ${resp.message}")
        }

        val respBody: InfoResponse = objectMapper.readValue(resp.body?.string()
            ?: throw IOException("Empty response"), InfoResponse::class.java)

        return respBody
    }
}