package gov.cdc.ocio.cosmossync.functions

import com.azure.cosmos.models.CosmosQueryRequestOptions
//import com.azure.cosmos.CosmosClient
import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpResponseMessage
import com.microsoft.azure.functions.HttpStatus
import gov.cdc.ocio.cosmossync.cosmos.CosmosClientManager
import gov.cdc.ocio.cosmossync.model.Item
import java.util.*
import java.util.logging.Logger
import com.azure.cosmos.CosmosClient

class HealthCheckFunction {

    fun run(
        request: HttpRequestMessage<Optional<String>>,
        context: ExecutionContext,
        cosmosClient: CosmosClient
    ): HttpStatus {
    
        try {
            //val cosmosClient = CosmosClientManager.getCosmosClient()

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