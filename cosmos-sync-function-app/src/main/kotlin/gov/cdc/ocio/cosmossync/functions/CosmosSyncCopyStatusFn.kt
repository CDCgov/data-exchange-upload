package gov.cdc.ocio.cosmossync.functions

import com.microsoft.azure.functions.annotation.*;
import com.microsoft.azure.functions.*
import gov.cdc.ocio.cosmossync.functions.CosmosSyncFunction
import gov.cdc.ocio.cosmossync.functions.HealthCheckFunction
import java.util.Optional

import com.google.gson.Gson
import gov.cdc.ocio.cosmossync.model.ItemCopyStatus
import com.azure.cosmos.CosmosContainer


class CosmosSyncCopyStatusFn {

    companion object {
        private const val containerName = "Items"
        private const val databaseName = "UploadStatus"
    } // .companion object 

    @FunctionName("CosmosSyncCopyStatusFn")
    fun evHubCopyStatus(
        @QueueTrigger(
            name = "msg",
            queueName = "%CosmosSinkQueueName%", 
            connection = "StorageConnectionString"
        ) message: String,
        context: ExecutionContext
    ) {
        val logger = context.logger

        logger.info("Dequeueing message: $message")
        val messageJson = Gson().fromJson(message, ItemCopyStatus::class.java)

        logger.info("Received JSON messageJson: ${messageJson}")

        // updateItemCopyStatus(context, container, messageJson)
    }

    private fun updateItemCopyStatus(context: ExecutionContext, container: CosmosContainer, msg: ItemCopyStatus) {

        val logger = context.logger

        logger.info("update id: ${msg.id}, statusDEX: ${msg.statusDEX}, statusEDAV: ${msg.statusEDAV}")
        
    } // .updateItemCopyStatus

  } // .CosmosSyncCopyStatusFn

  

