package gov.cdc.ocio.supplementalapi.functions

import com.azure.cosmos.models.CosmosQueryRequestOptions
import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpResponseMessage
import com.microsoft.azure.functions.HttpStatus
import gov.cdc.ocio.supplementalapi.cosmos.CosmosClientManager
import gov.cdc.ocio.supplementalapi.model.Item
import gov.cdc.ocio.supplementalapi.model.UploadStatus
import java.util.*


class StatusForTguidFunction {

    fun run(
        request: HttpRequestMessage<Optional<String>>,
        tguid: String,
        context: ExecutionContext
    ): HttpResponseMessage {

        val logger = context.logger

        logger.info("HTTP trigger processed a ${request.httpMethod.name} request.")

        val cosmosClient = CosmosClientManager.getCosmosClient()
        val cosmosDB = cosmosClient.getDatabase("UploadStatus")
        val container = cosmosDB.getContainer("Items")

        val sqlQuery = "select * from Items t where t.tguid = '$tguid'"
        val items = container.queryItems(
            sqlQuery, CosmosQueryRequestOptions(),
            Item::class.java
        )

        var uploadStatus: UploadStatus? = null
        if (items.iterator().hasNext()) {
            val item = items.iterator().next()
            uploadStatus = UploadStatus.createFromItem(item)
        }

        if (uploadStatus != null) {
            return request
                .createResponseBuilder(HttpStatus.OK)
                .header("Content-Type", "application/json")
                .body(uploadStatus)
                .build()
        }

        return request
            .createResponseBuilder(HttpStatus.BAD_REQUEST)
            .body("tguid '$tguid' not found")
            .build()
    }
}