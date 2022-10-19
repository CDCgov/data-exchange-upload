package gov.cdc.dataexchange.model

data class FileReceipt(var tguid: String? = null,
                       var etag: String? = null,
                       var status: String = "ACCEPTED")