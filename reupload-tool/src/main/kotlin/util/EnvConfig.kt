package org.example.util

import java.io.File
import java.util.*

class EnvConfig {
    companion object {
        private val propFile = File("local.properties")
        private val properties = if (propFile.exists()) Properties().apply {
            load(File("local.properties").inputStream())
        } else null
        val ENV: String = properties?.getProperty("dex.env") ?: System.getenv("DEX_ENV")
        val SAMS_USERNAME: String = properties?.getProperty("sams.username") ?: System.getenv("SAMS_USERNAME")
        val SAMS_PASSWORD: String = properties?.getProperty("sams.password") ?: System.getenv("SAMS_PASSWORD")
        val EDAV_STORAGE_ACCOUNT_NAME: String = properties?.getProperty("edav.storage.account.name") ?: System.getenv("EDAV_STORAGE_ACCOUNT_NAME")
        val EDAV_UPLOAD_CONTAINER_NAME: String = properties?.getProperty("edav.upload.container.name") ?: System.getenv("EDAV_UPLOAD_CONTAINER_NAME") ?: "upload"
        val ROUTING_STORAGE_CONNECTION_STRING: String = properties?.getProperty("routing.storage.connection.string") ?: System.getenv("ROUTING_STORAGE_CONNECTION_STRING")
        val ROUTING_UPLOAD_CONTAINER_NAME: String = properties?.getProperty("routing.upload.container.name") ?: System.getenv("ROUTING_UPLOAD_CONTAINER_NAME") ?: "routeingress"
        val AZURE_CLIENT_ID: String = properties?.getProperty("azure.client.id") ?: System.getenv("AZURE_CLIENT_ID")
        val AZURE_CLIENT_SECRET: String = properties?.getProperty("azure.client.secret") ?: System.getenv("AZURE_CLIENT_SECRET")
        val AZURE_TENANT_ID: String = properties?.getProperty("azure.tenant.id") ?: System.getenv("AZURE_TENANT_ID")
    }
}