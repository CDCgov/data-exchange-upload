package auth

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.KotlinModule
import okhttp3.FormBody
import okhttp3.OkHttpClient
import okhttp3.Request
import okio.IOException
import java.nio.charset.Charset
import java.nio.charset.StandardCharsets

class AuthClient(private val url: String) {
    private val httpClient = OkHttpClient()
    private val objectMapper = ObjectMapper().registerModule(KotlinModule())

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

        // TODO: Handle when body is empty.
        val respBody: AuthResponse = objectMapper.readValue(resp.body?.string() ?: "", AuthResponse::class.java)
        return respBody.accessToken
    }
}