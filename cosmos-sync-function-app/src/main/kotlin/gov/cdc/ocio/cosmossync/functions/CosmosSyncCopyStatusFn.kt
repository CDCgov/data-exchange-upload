package gov.cdc.ocio.cosmossync.functions

import com.microsoft.azure.functions.annotation.*;
import com.microsoft.azure.functions.*
import gov.cdc.ocio.cosmossync.functions.CosmosSyncFunction
import gov.cdc.ocio.cosmossync.functions.HealthCheckFunction
import java.util.Optional

import com.google.gson.Gson
import gov.cdc.ocio.cosmossync.model.*
import com.azure.cosmos.CosmosContainer
import com.azure.cosmos.models.PartitionKey
import gov.cdc.ocio.cosmossync.cosmos.CosmosClientManager


class CosmosSyncCopyStatus {

    companion object {
        private const val COSMOS_CONTAINER_NAME = "Items"
        private const val COSMOS_DB_NAME = "UploadStatus"
    } // .companion object 

    @FunctionName("CosmosSyncCopyStatusFn")
    fun evHubCopyStatus(
        @QueueTrigger(
            name = "msg",
            queueName = "%CosmosSinkCopyStatusQueueName%",  // cosmos-sink-copy-status-queue
            connection = "StorageConnectionString"
        ) message: String,
        context: ExecutionContext
    ) {

        val log = context.logger
        log.info("Dequeueing message: $message")

        try {

            val itemInternalStatus = Gson().fromJson(message, ItemInternalStatus::class.java)

            log.info("Received JSON itemInternalStatus: ${itemInternalStatus}")

            // cosmos connection    
            val cosmosClient = CosmosClientManager.getCosmosClient()
            val cosmosDb = cosmosClient.getDatabase(COSMOS_DB_NAME) 
            val cosmosContainer = cosmosDb.getContainer(COSMOS_CONTAINER_NAME)
            
            // get existing item from Cosmos by tguid
            val itemResponse = cosmosContainer.readItem(
                itemInternalStatus.tguid, PartitionKey(COSMOS_DB_NAME),
                ItemCopyStatus::class.java
            )
            val readItem: ItemCopyStatus = itemResponse.item

            // add the new internal statuses, two separated statuses in case of async updates
            readItem.statusDEX = itemInternalStatus.statusDEX
            readItem.statusEDAV = itemInternalStatus.statusEDAV

            // update the item
            cosmosContainer.upsertItem(readItem) 

        } catch (e: Exception) {

            log.info("Exception: ${e.localizedMessage}")

        } // .catch 

    } // .evHubCopyStatus


  } // .CosmosSyncCopyStatus

  

