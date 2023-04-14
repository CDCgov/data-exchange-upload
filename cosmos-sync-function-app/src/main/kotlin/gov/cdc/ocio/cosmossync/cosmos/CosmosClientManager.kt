package gov.cdc.ocio.cosmossync.cosmos

import com.azure.cosmos.ConsistencyLevel
import com.azure.cosmos.CosmosClient
import com.azure.cosmos.CosmosClientBuilder

class CosmosClientManager {
    companion object {

        private var client: CosmosClient? = null

        fun getCosmosClient(): CosmosClient {
            // Initialize a connection to cosmos that will persist across HTTP triggers
            val uri = System.getenv("CosmosDbEndpoint")
            val authKey = System.getenv("CosmosDbKey")

            if (client == null) {
                client = CosmosClientBuilder()
                    .endpoint(uri)
                    .key(authKey)
                    .consistencyLevel(ConsistencyLevel.EVENTUAL)
                    .contentResponseOnWriteEnabled(true)
                    .buildClient()
            }
            return client!!
        }
    }
}