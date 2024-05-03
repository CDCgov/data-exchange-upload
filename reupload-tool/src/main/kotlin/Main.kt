package org.example

import auth.AuthClient
import com.azure.identity.ClientSecretCredential
import com.azure.identity.ClientSecretCredentialBuilder
import com.azure.storage.blob.BlobServiceClient
import com.azure.storage.blob.BlobServiceClientBuilder
import com.github.doyaaaaaken.kotlincsv.dsl.csvReader
import org.example.model.Reupload
import org.example.util.EnvConfig
import tus.UploadClient
import java.io.File
import kotlin.collections.HashMap

fun main() {
    // First, read in input.csv.
    val inputCsv = getFileFromResources("input.csv")
    val reuploads = readInputCsv(inputCsv)
    println("Reuploading ${reuploads.size} file(s)")

    // Initialize auth client
    val dexUrl = when(EnvConfig.DEX_ENV) {
        "dev" -> "https://apidev.cdc.gov"
        "tst" -> "https://apitst.cdc.gov"
        "stg" -> "https://apistg.cdc.gov"
        "prd" -> "https://api.cdc.gov"
        else -> "https://apidev.cdc.gov"
    }
    val authClient = AuthClient(dexUrl)
    val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)

    // Initialize upload client
    val uploadClient = UploadClient(dexUrl, authToken)

    // Initialize blob service clients
    val edavBlobServiceClient = getBlobServiceClient(EnvConfig.EDAV_STORAGE_ACCOUNT_NAME,
        ClientSecretCredentialBuilder()
            .clientId(EnvConfig.AZURE_CLIENT_ID)
            .clientSecret(EnvConfig.AZURE_CLIENT_SECRET)
            .tenantId(EnvConfig.AZURE_TENANT_ID)
            .build())
    val routingBlobServiceClient = getBlobServiceClient(EnvConfig.ROUTING_STORAGE_CONNECTION_STRING)

    // Reupload files
    var successCount = 0
    var failCount = 0
    for (reupload: Reupload in reuploads) {
        try {

            // Check storage account
            val srcBlobClient = when (reupload.srcAccountId) {
                "edav" -> edavBlobServiceClient
                    .getBlobContainerClient(EnvConfig.EDAV_UPLOAD_CONTAINER_NAME)
                    .getBlobClient(reupload.src)

                "routing" -> routingBlobServiceClient
                    .getBlobContainerClient(EnvConfig.ROUTING_UPLOAD_CONTAINER_NAME)
                    .getBlobClient(reupload.src)

                else -> {
                    println("unsupported source storage account: ${reupload.srcAccountId}")
                    failCount++
                    continue
                }
            }

            // Download the file
            val srcFilename = srcBlobClient.blobName.split("/").last()
            println("downloading ${srcBlobClient.blobName} of size ${srcBlobClient.properties.blobSize}")
            srcBlobClient.downloadToFile("downloads/$srcFilename")

            // Overwrite filename in metadata with new filename
            val updatedMetadata = updateFilename(reupload.dest, srcBlobClient.properties.metadata)

            // Upload file.
            val fileToUpload = File("downloads/$srcFilename")
            println("uploading $fileToUpload")
            uploadClient.uploadFile(fileToUpload, updatedMetadata)
            successCount++
        } catch (e: Exception) {
            println("reupload failed")
            failCount++
        }
    }

    cleanupDownloads(File("downloads"))
    println("Reuploaded ${reuploads.size} files.  $successCount success. $failCount failed")
}

fun readInputCsv(inputFile: File): List<Reupload> {
    return csvReader().readAllWithHeader(inputFile).map {
        Reupload(
            it["src"]!!,
            it["dest"]!!,
            it["srcaccountid"]!!
        )
    }
}

fun getFileFromResources(filename: String): File {
    return File(object {}::class.java.classLoader.getResource(filename)?.file!!)
}

fun getBlobServiceClient(connectionString: String): BlobServiceClient {
    return BlobServiceClientBuilder()
        .connectionString(connectionString)
        .buildClient()
}

fun getBlobServiceClient(storageAccountName: String, azureCredentials: ClientSecretCredential): BlobServiceClient {
    return BlobServiceClientBuilder()
        .endpoint("https://$storageAccountName.blob.core.windows.net")
        .credential(azureCredentials)
        .buildClient()
}

// TODO: Make src immutable
fun updateFilename(filename: String, src: MutableMap<String, String>): MutableMap<String, String> {
    val res = src.toMutableMap()
    val filenameFields = listOf("filename", "original_filename", "meta_ext_filename", "received_filename")

    res.forEach{
        if (filenameFields.contains(it.key)) {
            res[it.key] = filename
        }
    }

    return res
}

fun cleanupDownloads(directory: File) {
    if (directory.exists() && directory.isDirectory) {
        directory.listFiles()?.forEach { file ->
            if (file.isDirectory) {
                cleanupDownloads(file)
            } else {
                file.delete()
            }
        }
    }
}