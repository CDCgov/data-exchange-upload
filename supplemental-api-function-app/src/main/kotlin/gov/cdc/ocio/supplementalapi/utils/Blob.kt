package gov.cdc.ocio.supplementalapi.utils

import com.azure.storage.blob.BlobClient
import java.io.ByteArrayOutputStream

object Blob {
    fun toByteArray(client: BlobClient): ByteArray {
        client.openInputStream().use { inputStream ->
            return inputStream.readBytes()
        }
    }
}