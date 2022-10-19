package gov.cdc.dataexchange.azure

import org.springframework.boot.context.properties.ConfigurationProperties
import org.springframework.context.annotation.Configuration

@ConfigurationProperties(value = "azure.blob")
@Configuration
class AzureBlobConfig {
    var connectStr: String? = null
    var containerName: String? = null
}