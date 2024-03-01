package util

import java.io.File
import java.util.Properties

class EnvConfig {
    companion object {
        private val propFile = File("local.properties")
        private val properties = if (propFile.exists()) Properties().apply {
            load(File("local.properties").inputStream())
        } else null
        val UPLOAD_URL: String = properties?.getProperty("upload.url") ?: System.getenv("UPLOAD_URL")
        val PROC_STAT_URL: String = properties?.getProperty("ps.api.url") ?: System.getenv("PS_API_URL")
        val SAMS_USERNAME: String = properties?.getProperty("sams.username") ?: System.getenv("SAMS_USERNAME")
        val SAMS_PASSWORD: String = properties?.getProperty("sams.password") ?: System.getenv("SAMS_PASSWORD")
        val DEX_STORAGE_CONNECTION_STRING: String = properties?.getProperty("dex.storage.connection.string") ?: System.getenv("DEX_STORAGE_CONNECTION_STRING")
    }
}