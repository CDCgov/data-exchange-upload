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
        val EDAV_STORAGE_ACCOUNT_NAME: String = properties?.getProperty("edav.storage.account.name") ?: System.getenv("EDAV_STORAGE_ACCOUNT_NAME")
        val ROUTING_STORAGE_CONNECTION_STRING: String = properties?.getProperty("routing.storage.connection.string") ?: System.getenv("ROUTING_STORAGE_CONNECTION_STRING")
        val AZURE_BLOB_SEARCH_DURATION_MILLIS: Long = properties?.getProperty("azure.blob.search.duration.millis")?.toLong() ?: System.getenv("AZURE_BLOB_SEARCH_DURATION_MILLIS")?.toLong() ?: 5_000
        val AZURE_CLIENT_ID: String = properties?.getProperty("azure.client.id") ?: System.getenv("AZURE_CLIENT_ID")
        val AZURE_CLIENT_SECRET: String = properties?.getProperty("azure.client.secret") ?: System.getenv("AZURE_CLIENT_SECRET")
        val AZURE_TENANT_ID: String = properties?.getProperty("azure.tenant.id") ?: System.getenv("AZURE_TENANT_ID")
    }
}