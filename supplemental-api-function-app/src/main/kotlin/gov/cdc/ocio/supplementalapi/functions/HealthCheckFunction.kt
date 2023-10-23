package gov.cdc.ocio.supplementalapi.functions

import com.azure.cosmos.models.CosmosQueryRequestOptions
import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpResponseMessage
import com.microsoft.azure.functions.HttpStatus
import gov.cdc.ocio.supplementalapi.cosmos.CosmosClientManager
import gov.cdc.ocio.supplementalapi.model.Item
import java.util.*
import com.azure.cosmos.CosmosClient
import mu.KotlinLogging


class HealthCheckFunction {
   
    fun run(
        request: HttpRequestMessage<Optional<String>>,
        context: ExecutionContext,
        cosmosClient: CosmosClient
    ): HttpStatus {     

        try {

            val logger = KotlinLogging.logger {} 
          
            println("---JSON LOG-START--")
            logger.info("logger-1 info")
            logger.warn("logger-2 warn")
            logger.error("logger-3 error")
            logger.debug("logger-4 debug")
            logger.trace("logger-5 trace")
            
            context.getLogger().info("Testing logs in azure");
            println("---JSON LOG-END--")

            val databaseName = System.getenv("CosmosDbDatabaseName")
            val containerName = System.getenv("CosmosDbContainerName")

            val cosmosDB = cosmosClient.getDatabase(databaseName)
            val container = cosmosDB.getContainer(containerName)

            val sqlQuery = "select * from $containerName t OFFSET 0 LIMIT 1"
            val items = container.queryItems(
                sqlQuery, CosmosQueryRequestOptions(),
                Item::class.java
            )

           return HttpStatus.OK

        } catch (ex: Exception) {
            println("An error occurred: ${ex.message}")

            return HttpStatus.INTERNAL_SERVER_ERROR
        }
    }
}