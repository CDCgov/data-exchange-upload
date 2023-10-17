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
            
            val databaseName = System.getenv("CosmosDbDatabaseName")
            val containerName = System.getenv("CosmosDbContainerName")

            val itemInternalCopyStatus = Gson().fromJson(message, ItemInternalCopyStatus::class.java)

            log.info("Received JSON itemInternalCopyStatus: ${itemInternalCopyStatus.toString()}")

            // cosmos connection    
            val cosmosClient = CosmosClientManager.getCosmosClient()
            val cosmosDb = cosmosClient.getDatabase(databaseName) 
            val cosmosContainer = cosmosDb.getContainer(containerName)

            // get existing item from Cosmos by tguid
            val itemResponse = cosmosContainer.readItem(
                itemInternalCopyStatus.tguid, PartitionKey(databaseName),
                ItemCopyStatus::class.java
            )
            val readItem: ItemCopyStatus = itemResponse.item
            log.info("Retrieved readItem.id: ${readItem.id}")

            // add the new internal statuses, two separated statuses in case of async updates
            readItem.statusDEX = itemInternalCopyStatus.statusDEX
            readItem.statusEDAV = itemInternalCopyStatus.statusEDAV

            log.info("Update readItem: $readItem")

            // update the item
            val response = cosmosContainer.upsertItem(readItem)

            log.info("Upsert success for tguid ${readItem.tguid}, new offset=${response.item.offset}") 

        } catch (e: Exception) {

            log.info("Exception: ${e.localizedMessage}")

        } // .catch 

    } // .evHubCopyStatus


  } // .CosmosSyncCopyStatus

  

