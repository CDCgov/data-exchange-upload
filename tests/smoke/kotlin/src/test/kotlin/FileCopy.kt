import auth.AuthClient
import com.azure.identity.ClientSecretCredentialBuilder
import com.azure.storage.blob.BlobClient
import com.azure.storage.blob.BlobContainerClient
import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import model.UploadConfig
import org.joda.time.DateTime
import org.joda.time.DateTimeZone
import org.testng.Assert
import org.testng.ITestContext
import org.testng.TestNGException
import org.testng.annotations.*
import org.testng.annotations.Optional
import tus.UploadClient
import util.*
import java.io.FileInputStream
import java.util.*
import kotlin.collections.HashMap
import kotlin.math.exp

@Listeners(UploadIdTestListener::class)
@Test()
class FileCopy {
    private val testFile = TestFile.getTestFileFromResources("10KB-test-file")
    private val authClient = AuthClient(EnvConfig.UPLOAD_URL)
    private val dexBlobClient = Azure.getBlobServiceClient(EnvConfig.DEX_STORAGE_CONNECTION_STRING)
    private val edavBlobClient = Azure.getBlobServiceClient(EnvConfig.EDAV_STORAGE_ACCOUNT_NAME,
        ClientSecretCredentialBuilder()
            .clientId(EnvConfig.AZURE_CLIENT_ID)
            .clientSecret(EnvConfig.AZURE_CLIENT_SECRET)
            .tenantId(EnvConfig.AZURE_TENANT_ID)
            .build())
    private val routingBlobClient = Azure.getBlobServiceClient(EnvConfig.ROUTING_STORAGE_CONNECTION_STRING)
    private lateinit var bulkUploadsContainerClient: BlobContainerClient
    private lateinit var uploadConfigBlobClient: BlobClient
    private lateinit var uploadConfig: UploadConfig
    private lateinit var dexContainerClient:BlobContainerClient
    private lateinit var edavContainerClient: BlobContainerClient
    private lateinit var routingContainerClient: BlobContainerClient
    private lateinit var uploadClient: UploadClient
    private lateinit var uploadId: String
    private lateinit var useCase: String
    private lateinit var metadata: HashMap<String, String>
    private lateinit var metadataMapping: HashMap<String, String>

    @Parameters("SENDER_MANIFEST", "USE_CASE", "SENDER_MANIFEST_MAPPING")
    @BeforeTest(groups = [Constants.Groups.FILE_COPY])
    fun beforeTest(
        context: ITestContext,
        @Optional("dextesting-testevent1.properties") SENDER_MANIFEST: String,
        @Optional("dextesting-testevent1") USE_CASE: String,
        @Optional("dextesting-testevent1-mapping.properties") SENDER_MANIFEST_MAPPING: String
    ) {
        useCase = USE_CASE

        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)

        val propertiesFilePath= "properties/$USE_CASE/$SENDER_MANIFEST"
        val propertiesFilePathMapping= "properties/$USE_CASE/$SENDER_MANIFEST_MAPPING"
        println("Storing Metadata")
        metadata = Metadata.convertPropertiesToMetadataMap(propertiesFilePath)
        metadataMapping = Metadata.convertPropertiesToMetadataMap(propertiesFilePathMapping)
        println(metadata)
        println(metadata.keys)
        println(metadataMapping)
        println(metadataMapping.keys)
        println("After Metadata")

        bulkUploadsContainerClient = dexBlobClient.getBlobContainerClient(Constants.BULK_UPLOAD_CONTAINER_NAME)
        println("dexBlobClient: $dexBlobClient.properties" )
        uploadConfigBlobClient = dexBlobClient
            .getBlobContainerClient(Constants.UPLOAD_CONFIG_CONTAINER_NAME)
            .getBlobClient("v1/${USE_CASE}.json")

        uploadConfig = jacksonObjectMapper().readValue(uploadConfigBlobClient.downloadContent().toString())

        dexContainerClient = dexBlobClient.getBlobContainerClient(USE_CASE)
        edavContainerClient = edavBlobClient.getBlobContainerClient(Constants.EDAV_UPLOAD_CONTAINER_NAME)
        routingContainerClient = routingBlobClient.getBlobContainerClient(Constants.ROUTING_UPLOAD_CONTAINER_NAME)
        uploadId = uploadClient.uploadFile(testFile, metadata) ?: throw TestNGException("Error uploading file ${testFile.name}")
        context.setAttribute("uploadId", uploadId)
        Thread.sleep(500) // Hard delay to wait for file to copy.

        Assert.assertTrue(bulkUploadsContainerClient.exists())
    }

    @Test(groups = [Constants.Groups.FILE_COPY])
    fun shouldUploadFileToTusContainer() {
        val uploadBlob = bulkUploadsContainerClient.getBlobClient("${Constants.TUS_PREFIX_DIRECTORY_NAME}/$uploadId")
        val uploadInfoBlob = bulkUploadsContainerClient.getBlobClient("${Constants.TUS_PREFIX_DIRECTORY_NAME}/$uploadId.info")

        Assert.assertTrue(uploadBlob.exists())
        Assert.assertTrue(uploadInfoBlob.exists())
    }

    @Test(groups = [Constants.Groups.FILE_COPY])
    fun shouldHaveSameSizeFileInTusContainer() {
        val uploadBlob = bulkUploadsContainerClient.getBlobClient("${Constants.TUS_PREFIX_DIRECTORY_NAME}/$uploadId")

        Assert.assertEquals(uploadBlob.properties.blobSize, testFile.length())
    }

    @Test(groups = [Constants.Groups.FILE_COPY])
    fun shouldCopyToDestinationContainers() {
        val filenameSuffix = if (uploadConfig.copyConfig.filenameSuffix == "upload_id") "_${uploadId}" else ""
        val expectedFilename = "${Metadata.getFilePrefixByDate(DateTime(DateTimeZone.UTC), useCase)}/${testFile.nameWithoutExtension}${filenameSuffix}${testFile.extension}"
        var expectedBlobClient: BlobClient?

        if (uploadConfig.copyConfig.targets.contains("edav")) {
            expectedBlobClient = edavContainerClient.getBlobClient(expectedFilename)

            Assert.assertNotNull(expectedBlobClient)
            Assert.assertEquals(expectedBlobClient!!.properties.blobSize, testFile.length())
        }

        if (uploadConfig.copyConfig.targets.contains("routing")) {
            expectedBlobClient = routingContainerClient.getBlobClient(expectedFilename)

            Assert.assertNotNull(expectedBlobClient)
            Assert.assertEquals(expectedBlobClient!!.properties.blobSize, testFile.length())
        }
    }

    @Test(groups = [Constants.Groups.FILE_COPY])
    fun shouldTranslateMetadataGivenV1SenderManifest() {

        val filenameSuffix = if (uploadConfig.copyConfig.filenameSuffix == "upload_id") "_${uploadId}" else ""
        val expectedFilename = "${Metadata.getFilePrefixByDate(DateTime(DateTimeZone.UTC), useCase)}/${testFile.nameWithoutExtension}$filenameSuffix.${testFile.extension}"
        val modifiedFilename = expectedFilename.removePrefix("$useCase/")

        val expectedBlobClient = dexContainerClient.getBlobClient(modifiedFilename)

        val blobProperties = expectedBlobClient?.properties
            ?: throw AssertionError("Blob client has no properties.")

        val blobMetadata = blobProperties.metadata

        //Metadata Value Validation
        metadata.forEach { (v1Key, v1Value) ->
            val expectedFieldInV2 = metadataMapping[v1Key]
            val actualValueInV2 = blobMetadata[expectedFieldInV2]
            Assert.assertEquals(v1Value, actualValueInV2, "Expected V1 key value: $v1Value does not match with actual V2 key value: $actualValueInV2")
        }

        //Metadata Key Validation
        metadataMapping.forEach { (v1Key, v2Key) ->
            Assert.assertTrue(blobMetadata.containsKey(v2Key), "Mismatch:Blob metadata does not contain expected V2 key: $v2Key which should map from V1 key: $v1Key")
        }
    }
}
