package util

import java.io.File
import java.util.Properties

class EnvConfig {
    companion object {
        private val propFile = File("local.properties")
        private val properties = if (propFile.exists()) Properties().apply {
            load(File("local.properties").inputStream())
        } else null
        public final val ENVIRONMENT: Environment = Environment.valueOf(properties?.getProperty("environment") ?: System.getenv("ENVIRONMENT"))
        val UPLOAD_URL: String = properties?.getProperty("upload.url") ?: System.getenv("UPLOAD_URL")
        val PROC_STAT_URL: String = properties?.getProperty("ps.api.url") ?: System.getenv("PS_API_URL")
        val SAMS_USERNAME: String = properties?.getProperty("sams.username") ?: System.getenv("SAMS_USERNAME")
        val SAMS_PASSWORD: String = properties?.getProperty("sams.password") ?: System.getenv("SAMS_PASSWORD")
        val HEALTHCHECK_CASE: String = properties?.getProperty("healthcheck.case") ?: System.getenv("HEALTHCHECK_CASE")
    }
}
