package gov.cdc.dataexchange.azure

import com.azure.storage.blob.BlobClient
import com.azure.storage.blob.BlobContainerClient
import com.azure.storage.blob.BlobServiceClient
import com.azure.storage.blob.BlobServiceClientBuilder
import com.azure.storage.blob.models.BlobHttpHeaders
import gov.cdc.dataexchange.model.FileReceipt
import org.slf4j.LoggerFactory
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.stereotype.Component
import org.springframework.web.multipart.MultipartFile
import java.io.IOException
import java.nio.file.Files
import java.nio.file.Path
import java.text.SimpleDateFormat
import java.util.*

@Component
class BlobProxy {

    private val MD_SOURCE_TIMESTAMP = "meta_ext_sourcetimestamp"

    private val formatter = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:sss")

    private val log = LoggerFactory.getLogger(BlobProxy::class.java)

    @Autowired
    var blobConfig: AzureBlobConfig? = null

    private fun initClient(): BlobContainerClient {
        val blobServiceClient = BlobServiceClientBuilder()
            .connectionString(blobConfig?.connectStr)
            .buildClient()
        log.info("connectStr: {}", blobConfig?.connectStr)
        log.info("containerName: {}", blobConfig?.containerName)
        return blobServiceClient.getBlobContainerClient(blobConfig?.containerName)
    }

    private fun getBlobClient(key: String?): BlobClient {
        val blobContainerClient = initClient()
        return blobContainerClient.getBlobClient(key)
    }

    @Throws(IOException::class)
    fun uploadFile(
        file: MultipartFile,
        parameters: MutableMap<String?, String?>
    ): FileReceipt {
        val tguid = UUID.randomUUID().toString()
        parameters[MD_SOURCE_TIMESTAMP] = formatter.format(Date())
        val headers = BlobHttpHeaders().setContentType(file.contentType)
        val path = Files.createFile(Path.of("/tmp", tguid))
        file.transferTo(path)
        val blobClient = getBlobClient(path.fileName.toString())
        blobClient.uploadFromFile(path.toString(), null, headers, parameters, null, null, null)
        Files.delete(path)
        return FileReceipt(
            tguid,
            if (parameters["meta_ext_filename"] != null) parameters["meta_ext_filename"] else "nofilename"
        )
    }

}