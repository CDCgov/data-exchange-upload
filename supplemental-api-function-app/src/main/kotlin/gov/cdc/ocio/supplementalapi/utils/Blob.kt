package gov.cdc.ocio.supplementalapi.utils

import com.azure.storage.blob.BlobClient
import java.io.ByteArrayOutputStream

object Blob {
    fun toByteArray(client: BlobClient): ByteArray {
        ByteArrayOutputStream().use { outputStream ->
            client.downloadStream(outputStream)
            return outputStream.toByteArray()
        }
        // TODO: Throw exception if we get here.
    }
}