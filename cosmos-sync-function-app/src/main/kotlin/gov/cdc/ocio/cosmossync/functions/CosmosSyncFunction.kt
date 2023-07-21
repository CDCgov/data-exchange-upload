package gov.cdc.ocio.cosmossync.functions

import com.azure.cosmos.CosmosClient
import com.azure.cosmos.CosmosContainer
import com.azure.cosmos.CosmosDatabase
import com.azure.cosmos.CosmosException
import com.azure.cosmos.models.*
import com.google.gson.Gson
import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.annotation.FunctionName
import com.microsoft.azure.functions.annotation.HttpTrigger
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpResponseMessage
import com.microsoft.azure.functions.HttpStatus
import com.microsoft.azure.functions.HttpMethod
import gov.cdc.ocio.cosmossync.cosmos.CosmosClientManager
import gov.cdc.ocio.cosmossync.model.Item
import java.util.*

class CosmosSyncFunction {

    fun run(
        context: ExecutionContext,
        message: String
    ) {
        context.logger.info("Dequeueing message: $message")
        val update = Gson().fromJson(message, Item::class.java)

        postReceive(context, update)
    }

    private fun postReceive(context: ExecutionContext, update: Item) {
        val logger = context.logger

        logger.info("tguid = ${update.tguid}, offset = ${update.offset}, size = ${update.size}")

        logger.info("calling initDatabaseContainer...")
        val container = initDatabaseContainer(context, "Items")!!
        logger.info("done calling initDatabaseContainer")

        upsertItem(context, container, update)
    }

    @Throws(Exception::class)
    private fun createDatabaseIfNotExists(context: ExecutionContext, cosmosClient: CosmosClient, databaseName: String): CosmosDatabase? {
        context.logger.info("Create database $databaseName if not exists...")

        //  Create database if not exists
        val databaseResponse = cosmosClient.createDatabaseIfNotExists(databaseName)
        return cosmosClient.getDatabase(databaseResponse.properties.id)
    }

    private fun initDatabaseContainer(context: ExecutionContext, containerName: String): CosmosContainer? {
        try {
            val logger = context.logger

            logger.info("calling getCosmosClient...")
            val cosmosClient = CosmosClientManager.getCosmosClient()

            // setup database
            logger.info("calling createDatabaseIfNotExists...")
            val db = createDatabaseIfNotExists(context, cosmosClient, "UploadStatus")!!

            val containerProperties = CosmosContainerProperties(containerName, "/partitionKey")

            // Provision throughput
            val throughputProperties = ThroughputProperties.createManualThroughput(400)

            //  Create container with 400 RU/s
            logger.info("calling createContainerIfNotExists...")
            val databaseResponse = db.createContainerIfNotExists(containerProperties, throughputProperties)

            return db.getContainer(databaseResponse.properties.id)

        } catch (ex: CosmosException) {
            context.logger.info("exception: ${ex.localizedMessage}")
        }
        return null
    }

    private fun upsertItem(context: ExecutionContext, container: CosmosContainer, update: Item) {
        val logger = context.logger

        logger.info("Upserting tguid = ${update.tguid}")

        logger.info("tguid: ${update.tguid}")
        logger.info("offset: ${update.offset}")
        logger.info("size: ${update.size}")
/*
        # latest_offset = 0
        # with open('/tmp/{0}.txt'.format(tguid)) as f:
        #     latest_offset = int(f.readline())
        #     f.close()

        # logger.info('upsert_item: latest_offset = {0}, our offset = {1}'.format(latest_offset, offset))
        # if (offset < latest_offset):
        #     # This update is stale so skip it.  This is due to threading as order of upsert_item
        #     # calls at this point is no longer guaranteed.
        #     logger.warning("Our information is stale - skipping (latest_offset = {0}, this offset = {1})".format(latest_offset, offset))
        #     return
*/
        var readItem: Item? = null
        try {
            logger.info("Checking to see if tguid exists...")
            val itemResponse = container.readItem(
                update.tguid, PartitionKey("UploadStatus"),
                Item::class.java
            )
            readItem = itemResponse.item

            logger.info("Found tguid with previous offset = ${readItem.offset}")
            if (readItem.offset >= update.offset) {
                logger.warning("Out of order call, continuing...")
                return
            }
            logger.info("Updating found tguid with new offset")
            readItem.offset = update.offset

        } catch (ex: CosmosException) {
            // If here, the tguid was not found
        }

        if (readItem == null) {
            logger.info("*** tguid not found")
            readItem = Item()
            readItem.id = update.tguid
            readItem.tguid = update.tguid
            readItem.partitionKey = "UploadStatus"
            readItem.filename = update.filename
            readItem.offset = update.offset
            readItem.size = update.size
            readItem.meta_destination_id = update.meta_destination_id
            readItem.meta_ext_event = update.meta_ext_event
            readItem.metadata = update.metadata
        }

        logger.info("Calling upsert_item")
        val response = container.upsertItem(readItem)
        logger.info("Done calling upsert_item")

        logger.info("Upserted at ${Date()}, new offset=${response.item.offset}")
        logger.info("Upsert success for tguid ${update.tguid}!!")
    }

    @FunctionName("CosmosHealthCheck")
    fun healthCheck(
        @HttpTrigger(
            name = "req",
            methods = [HttpMethod.GET],
            route = "health"
        ) request: HttpRequestMessage<Optional<String>>,
        context: ExecutionContext
    ): HttpResponseMessage {
        // Perform your health check logic here
        val isHealthy = performHealthCheck(context)

        // Return appropriate response based on health status
        return if (isHealthy) {
            request.createResponseBuilder(HttpStatus.OK).body("CosmosSyncFunction is healthy").build()
        } else {
            request.createResponseBuilder(HttpStatus.INTERNAL_SERVER_ERROR).body("CosmosSyncFunction is not healthy").build()
        }
    }

    private fun performHealthCheck(context: ExecutionContext): Boolean {
        try {
            val logger = context.logger

            // Get the Cosmos DB client
            logger.info("Checking Cosmos DB client availability...")
            val cosmosClient: CosmosClient? = CosmosClientManager.getCosmosClient()
            if (cosmosClient != null) {
                // If the client is available, consider the function healthy
                return true
            } else {
                // If the client is null, there is an issue with Cosmos DB connectivity
                // Log the error or handle it based on your application's requirements
                logger.warning("Cosmos DB client is not available.")
                return false
            }

        } catch (ex: Exception) {
            context.logger.warning("Health check failed: ${ex.localizedMessage}")
            return false
        }
    }
}