import com.azure.storage.blob.BlobContainerClient
import io.tus.java.client.ProtocolException
import dex.DexUploadClient
import model.UploadConfig
import org.joda.time.DateTime
import org.joda.time.DateTimeZone
import org.testng.Assert
import org.testng.ITestContext
import org.testng.annotations.*
import tus.UploadClient
import util.*
import util.ConfigLoader.Companion.loadUploadConfig
import util.Constants.Companion.TEST_DESTINATION
import util.Constants.Companion.TEST_EVENT
import util.DataProvider

@Listeners(UploadIdTestListener::class)
@Test()
class MetadataVerify {
    private val testFile = TestFile.getTestFileFromResources("10KB-test-file")
    private val authClient = DexUploadClient(EnvConfig.UPLOAD_URL)
    private lateinit var authToken: String
    private lateinit var uploadClient: UploadClient
    private lateinit var metadata: HashMap<String, String>
    private val dexBlobClient = Azure.getBlobServiceClient(EnvConfig.DEX_STORAGE_CONNECTION_STRING)
    private lateinit var dexContainerClient: BlobContainerClient
    private lateinit var useCase: String
    private lateinit var senderManifest: String
    private lateinit var senderManifestInvalidFilename: String
    private lateinit var senderManifestNoDestId: String
    private lateinit var senderManifestNoEvent: String
    private lateinit var uploadConfigV1: UploadConfig
    private lateinit var uploadConfigV2: UploadConfig

    @Parameters(
        "USE_CASE",
        "SENDER_MANIFEST",
        "SENDER_MANIFEST_INVALID_FILENAME",
        "SENDER_MANIFEST_NO_DEST_ID",
        "SENDER_MANIFEST_NO_EVENT"
    )
    @BeforeTest(groups = [Constants.Groups.METADATA_VERIFY])
    fun beforeTest(
        @Optional("dextesting-testevent1") USE_CASE: String,
        @Optional("dextesting-testevent1.properties") SENDER_MANIFEST: String,
        @Optional("invalid-filename.properties") SENDER_MANIFEST_INVALID_FILENAME: String,
        @Optional("no-dest-id.properties") SENDER_MANIFEST_NO_DEST_ID: String,
        @Optional("no-event.properties") SENDER_MANIFEST_NO_EVENT: String
    ) {
        authToken = authClient.getToken(EnvConfig.SAMS_USERNAME, EnvConfig.SAMS_PASSWORD)

        useCase = USE_CASE
        senderManifest = SENDER_MANIFEST
        senderManifestInvalidFilename = SENDER_MANIFEST_INVALID_FILENAME
        senderManifestNoDestId = SENDER_MANIFEST_NO_DEST_ID
        senderManifestNoEvent = SENDER_MANIFEST_NO_EVENT

        uploadConfigV1 = loadUploadConfig(dexBlobClient, USE_CASE, "v1")
        uploadConfigV2 = loadUploadConfig(dexBlobClient, USE_CASE, "v2")
        dexContainerClient = dexBlobClient.getBlobContainerClient(useCase)

    }

    @BeforeMethod
    fun beforeMethod() {
        uploadClient = UploadClient(EnvConfig.UPLOAD_URL, authToken)
    }

    @Test(
        groups = [Constants.Groups.METADATA_VERIFY],
        dataProvider = "versionProvider", dataProviderClass = DataProvider::class
    )
    fun shouldUploadFileGivenRequiredMetadata(context: ITestContext, version: String) {
        metadata = Metadata.getMetadataMap(version, useCase, senderManifest)
        val uploadId = uploadClient.uploadFile(testFile, metadata)
        context.setAttribute("uploadId_$version", uploadId)
        Assert.assertNotNull(uploadId)
    }

    @Test(
        groups = [Constants.Groups.METADATA_VERIFY],
        expectedExceptions = [ProtocolException::class],
        expectedExceptionsMessageRegExp = "unexpected status code \\(400\\).*",
        dataProvider = "versionProvider", dataProviderClass = DataProvider::class
    )
    fun shouldReturnErrorWhenDestinationIDNotProvided(version: String) {
        metadata = Metadata.getMetadataMap(version, useCase, senderManifestNoDestId)
        uploadClient.uploadFile(testFile, metadata)
    }

    @Test(
        groups = [Constants.Groups.METADATA_VERIFY],
        expectedExceptions = [ProtocolException::class],
        expectedExceptionsMessageRegExp = "unexpected status code \\(400\\).*",
        dataProvider = "versionProvider", dataProviderClass = DataProvider::class
    )
    fun shouldReturnErrorWhenEventNotProvided(version: String) {
        metadata = Metadata.getMetadataMap(version, useCase, senderManifestNoEvent)
        uploadClient.uploadFile(testFile, metadata)
    }

    @Test(
        groups = [
            Constants.Groups.METADATA_VERIFY],
        expectedExceptions = [ProtocolException::class],
        expectedExceptionsMessageRegExp = "unexpected status code \\(400\\).*field filename was missing"
    )
    fun shouldReturnErrorWhenFilenameNotProvided() {
        val metadata = hashMapOf(
            "meta_destination_id" to TEST_DESTINATION,
            "meta_ext_event" to TEST_EVENT,
            "meta_ext_source" to "INTEGRATION-TEST"
        )
        uploadClient.uploadFile(testFile, metadata)
    }

    @Test(
        groups = [Constants.Groups.METADATA_VERIFY],
        expectedExceptions = [ProtocolException::class],
        expectedExceptionsMessageRegExp = "unexpected status code \\(400\\).*",
        dataProvider = "versionProvider", dataProviderClass = DataProvider::class
    )
    fun shouldReturnErrorWhenFilenameContainsInvalidChars(version: String) {
        metadata = Metadata.getMetadataMap(version, useCase, senderManifestInvalidFilename)
        uploadClient.uploadFile(testFile, metadata)
    }

    @Test(
        groups = [Constants.Groups.METADATA_VERIFY],
        dataProvider = "versionProvider", dataProviderClass = DataProvider::class
    )
    fun shouldValidateV2MetadataWithSenderManifest(version: String) {

        metadata = Metadata.getMetadataMap(version, useCase, senderManifest)
        val uploadId = uploadClient.uploadFile(testFile, metadata)

        val uploadConfig = if (version == "V1") uploadConfigV1 else uploadConfigV2
        val metadataFields = uploadConfig.metadataConfig.fields


        Thread.sleep(500)
        val filenameSuffix = if (uploadConfig.copyConfig.filenameSuffix == "upload_id") "_${uploadId}" else ""

        val expectedFilename =
            "${Metadata.getFilePrefixByDate(DateTime(DateTimeZone.UTC))}/${testFile.nameWithoutExtension}$filenameSuffix.${testFile.extension}"

        val expectedBlobClient = dexContainerClient.getBlobClient(expectedFilename)

        val blobProperties = expectedBlobClient.properties
        val blobMetadata = blobProperties.metadata

        metadata.forEach { (key, value) ->
            val actualValue = blobMetadata[key]
            Assert.assertEquals(
                value, actualValue, "Expected key value: $value does not match with actual key value: $actualValue"
            )
        }
        metadataFields.forEach { field ->
            Assert.assertTrue(blobMetadata.containsKey(field.fieldName), "V2 keys mismatch: ${field.fieldName}")
        }
    }
}