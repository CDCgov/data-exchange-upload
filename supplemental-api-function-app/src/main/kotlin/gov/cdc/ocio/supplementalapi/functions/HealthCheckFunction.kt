package gov.cdc.ocio.supplementalapi.functions

import com.azure.cosmos.CosmosClient
import com.azure.cosmos.CosmosClientBuilder
import com.azure.cosmos.models.CosmosQueryRequestOptions
import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpResponseMessage
import com.microsoft.azure.functions.HttpStatus
import gov.cdc.ocio.supplementalapi.model.Item
import java.util.*


class HealthCheckFunction {
    private val cosmosClient: CosmosClient
    // Initialize the CosmosDB client in the constructor
    init {
        val endpoint = System.getenv("COSMOSDBENDPOINT")
        val cosmoDBKey = System.getenv("COSMOSDBKEY")

        val cosmosClientBuilder = CosmosClientBuilder()
            .endpoint(endpoint)
            .key(cosmoDBKey)
        cosmosClient = cosmosClientBuilder.buildClient()
    }

    fun run(
        request: HttpRequestMessage<Optional<String>>,
        context: ExecutionContext
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

           /* return request.createResponseBuilder(HttpStatus.OK)
                   .body("Cosmos DB is healthy")
                   .build()*/

            return HttpStatus.OK


        } catch (ex: Throwable) {
            println("An error occurred: ${ex.message}")

            return HttpStatus.INTERNAL_SERVER_ERROR

            /*return request.createResponseBuilder(HttpStatus.OK)
                .body("Cosmos DB not healthy")
                .build()*/
        }
    }
}