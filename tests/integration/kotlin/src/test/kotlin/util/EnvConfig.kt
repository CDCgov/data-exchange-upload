package util

import java.io.File
import java.util.Properties

class EnvConfig {
    companion object {
        private val properties = Properties().apply {
            load(File("local.properties").inputStream())
        }
        val UPLOAD_URL: String = properties.getProperty("upload.url") ?: System.getenv("UPLOAD_URL")
        val PROC_STAT_URL: String = properties.getProperty("ps.api.url") ?: System.getenv("PS_API_URL")
        val SAMS_USERNAME: String = properties.getProperty("sams.username") ?: System.getenv("SAMS_USERNAME")
        val SAMS_PASSWORD: String = properties.getProperty("sams.password") ?: System.getenv("SAMS_PASSWORD")
    }
}