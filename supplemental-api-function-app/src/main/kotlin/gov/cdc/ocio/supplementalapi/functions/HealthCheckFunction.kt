package gov.cdc.ocio.supplementalapi.functions

import com.azure.cosmos.models.CosmosQueryRequestOptions
import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpResponseMessage
import com.microsoft.azure.functions.HttpStatus
import gov.cdc.ocio.supplementalapi.cosmos.CosmosClientManager
import gov.cdc.ocio.supplementalapi.model.Item
import java.util.*


class HealthCheckFunction() {

    fun run(
        request: HttpRequestMessage<Optional<String>>,
        context: ExecutionContext
    ): HttpResponseMessage {

        try {

            val cosmosClient = CosmosClientManager.getCosmosClient()

            val databaseName = System.getenv("CosmosDbDatabaseName")
            val containerName = System.getenv("CosmosDbContainerName")

            val cosmosDB = cosmosClient.getDatabase(databaseName)
            val container = cosmosDB.getContainer(containerName)

            val sqlQuery = "select * from $containerName t OFFSET 0 LIMIT 1"
            val items = container.queryItems(
                sqlQuery, CosmosQueryRequestOptions(),
                Item::class.java
            )

           return request.createResponseBuilder(HttpStatus.OK)
                   .body("Cosmos DB is healthy")
                   .build()


        } catch (ex: Throwable) {
            println("An error occurred: ${ex.message}")

            return request.createResponseBuilder(HttpStatus.OK)
                .body("Cosmos DB not healthy")
                .build()
        }
    }
}