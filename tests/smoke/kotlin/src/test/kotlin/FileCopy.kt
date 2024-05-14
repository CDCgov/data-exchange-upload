import auth.AuthClient
import com.azure.identity.ClientSecretCredentialBuilder
import com.azure.storage.blob.BlobClient
import com.azure.storage.blob.BlobContainerClient
import model.ConfigLoader
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
import kotlin.collections.HashMap


@Listeners(UploadIdTestListener::class)
@Test()
class FileCopy {
    private val testFile = TestFile.getTestFileFromResources("10KB-test-file")
    private val authClient = AuthClient(EnvConfig.UPLOAD_URL)
    private val dexBlobClient = Azure.getBlobServiceClient(EnvConfig.DEX_STORAGE_CONNECTION_STRING)
    private val edavBlobClient = Azure.getBlobServiceClient(
        EnvConfig.EDAV_STORAGE_ACCOUNT_NAME,
        ClientSecretCredentialBuilder()
            .clientId(EnvConfig.AZURE_CLIENT_ID)
            .clientSecret(EnvConfig.AZURE_CLIENT_SECRET)
            .tenantId(EnvConfig.AZURE_TENANT_ID)
            .build()
    )
    private val routingBlobClient = Azure.getBlobServiceClient(EnvConfig.ROUTING_STORAGE_CONNECTION_STRING)
    private lateinit var bulkUploadsContainerClient: BlobContainerClient
    private lateinit var dexContainerClient: BlobContainerClient
    private lateinit var edavContainerClient: BlobContainerClient
    private lateinit var routingContainerClient: BlobContainerClient
    private lateinit var uploadClient: UploadClient
    private lateinit var configLoader: ConfigLoader
    private lateinit var uploadId: String
    private lateinit var useCase: String
    private lateinit var uploadConfigV1: UploadConfig
    private lateinit var uploadConfigV2: UploadConfig
    private lateinit var metadata: HashMap<String, String>
    private lateinit var metadataV2: HashMap<String, String>

    @Parameters("SENDER_MANIFEST", "USE_CASE", "USE_CASEV2")
    @BeforeTest(groups = [Constants.Groups.FILE_COPY])
    fun beforeTest(
        context: ITestContext,
        @Optional("dextesting-testevent1.properties") SENDER_MANIFEST: String,
        @Optional("dextesting-testevent1") USE_CASE: String,
        @Optional("dextesting-testevent1") USE_CASEV2: String
    ) {
        useCase = USE_CASE

        val authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)

        val propertiesFilePath = "properties/$USE_CASE/$SENDER_MANIFEST"
        val propertiesFilePathV2 = "properties/$USE_CASEV2/$SENDER_MANIFEST"

        metadata = Metadata.convertPropertiesToMetadataMap(propertiesFilePath)
        metadataV2 = Metadata.convertPropertiesToMetadataMap(propertiesFilePathV2)

        bulkUploadsContainerClient = dexBlobClient.getBlobContainerClient(Constants.BULK_UPLOAD_CONTAINER_NAME)
        println("dexBlobClient: $dexBlobClient.properties")

        configLoader = ConfigLoader()
        uploadConfigV1 = configLoader.loadUploadConfig(dexBlobClient, USE_CASE, "v1")
        uploadConfigV2 = configLoader.loadUploadConfig(dexBlobClient, USE_CASE, "v2")

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
        val uploadInfoBlob =
            bulkUploadsContainerClient.getBlobClient("${Constants.TUS_PREFIX_DIRECTORY_NAME}/$uploadId.info")

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
        val filenameSuffix = Filename.getFilenameSuffix(uploadConfigV1.copyConfig, uploadId)
        val expectedFilename = "${
            Metadata.getFilePrefixByDate(
                DateTime(DateTimeZone.UTC),
                useCase
            )
        }/${testFile.nameWithoutExtension}${filenameSuffix}${testFile.extension}"
        var expectedBlobClient: BlobClient?

        if (uploadConfigV1.copyConfig.targets.contains("edav")) {
            expectedBlobClient = edavContainerClient.getBlobClient(expectedFilename)

            Assert.assertNotNull(expectedBlobClient)
            Assert.assertEquals(expectedBlobClient!!.properties.blobSize, testFile.length())
        }

        if (uploadConfigV1.copyConfig.targets.contains("routing")) {
            expectedBlobClient = routingContainerClient.getBlobClient(expectedFilename)

            Assert.assertNotNull(expectedBlobClient)
            Assert.assertEquals(expectedBlobClient!!.properties.blobSize, testFile.length())
        }
    }

    @Test(groups = [Constants.Groups.FILE_COPY])
    fun shouldTranslateMetadataGivenV1SenderManifest() {
        val metadataMapping = uploadConfigV2.metadataConfig.fields
            .associate { it.compatFieldName to it.fieldName }

        val filenameSuffix = Filename.getFilenameSuffix(uploadConfigV1.copyConfig, uploadId)

        val expectedFilename =
            "${Metadata.getFilePrefixByDate(DateTime(DateTimeZone.UTC))}/${testFile.nameWithoutExtension}$filenameSuffix.${testFile.extension}"

        val expectedBlobClient = dexContainerClient.getBlobClient(expectedFilename)

        val blobProperties = expectedBlobClient.properties
        val blobMetadata = blobProperties.metadata

        metadataMapping.forEach { (v1Key, v2Key) ->
            Assert.assertTrue(blobMetadata.containsKey(v2Key), "Mismatch: Blob metadata does not contain expected V2 key: $v2Key which should map from V1 key: $v1Key")
        }

        metadata.forEach { (v1Key, v1Value) ->
            val expectedFieldInV2 = metadataMapping[v1Key]
            val actualValueInV2 = blobMetadata[expectedFieldInV2]
            Assert.assertEquals(v1Value, actualValueInV2, "Expected V1 key value: $v1Value does not match with actual V2 key value: $actualValueInV2")
        }
    }

    @Test(groups = [Constants.Groups.FILE_COPY])
    fun shouldValidateV2MetadataWithSenderManifest() {

        val metadataFields = uploadConfigV2.metadataConfig.fields

        val filenameSuffix = if (uploadConfigV1.copyConfig.filenameSuffix == "upload_id") "_${uploadId}" else ""
        val expectedFilename =
            "${Metadata.getFilePrefixByDate(DateTime(DateTimeZone.UTC))}/${testFile.nameWithoutExtension}$filenameSuffix.${testFile.extension}"

        val expectedBlobClient = dexContainerClient.getBlobClient(expectedFilename)

        val blobProperties = expectedBlobClient.properties
        val blobMetadata = blobProperties.metadata

        metadataV2.forEach { (key, value) ->
            val actualValueInV2 = blobMetadata[key]
            Assert.assertEquals(value, actualValueInV2, "Expected key value: $value does not match with actual key value: $actualValueInV2")
        }
        metadataFields.forEach { field  ->
            Assert.assertTrue(blobMetadata.containsKey(field.fieldName), "V2 keys mismatch: ${field.fieldName}")
        }
    }
}



