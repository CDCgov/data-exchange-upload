package gov.cdc.dataexchange.azure

import org.springframework.boot.context.properties.ConfigurationProperties
import org.springframework.context.annotation.Configuration
import org.springframework.beans.factory.annotation.Value;
import com.azure.identity.DefaultAzureCredentialBuilder
import com.azure.security.keyvault.secrets.SecretClientBuilder
import org.slf4j.LoggerFactory

@ConfigurationProperties(value = "azure.blob")
@Configuration
class AzureBlobConfig {

//    @Value("\${azure.blob.connectStr}")
    var connectStr: String? = null

    var containerName: String? = null

    private val log = LoggerFactory.getLogger(AzureBlobConfig::class.java)

    init {
        try {
            // Check if we can load from key vault
            val keyVaultName = "tf-ede-envar-vault"
            val keyVaultUri = "https://$keyVaultName.vault.azure.net"
            val secretName = "BULK-UPLOAD-INGRESS-BLOB-STORAGE-CONNECTION-STRING"

            log.info("key vault name = $keyVaultName and key vault URI = $keyVaultUri")

            val secretClient = SecretClientBuilder()
                .vaultUrl(keyVaultUri)
                .credential(DefaultAzureCredentialBuilder().build())
                .buildClient()

            log.info("Retrieving your secret from $keyVaultName.")
            val retrievedSecret = secretClient.getSecret(secretName)
            log.info("Your secret's value is '" + retrievedSecret.value + "'.")
        } catch (e: Exception) {
            log.error("Failed to retrieve key vault secret for blob connection string, using application defaults")
        }
    }
}