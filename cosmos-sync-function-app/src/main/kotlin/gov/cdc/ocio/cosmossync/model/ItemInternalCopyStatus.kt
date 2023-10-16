package gov.cdc.ocio.cosmossync.model

// sent by BulkFileUploadFunctionApp to storage queue 
class ItemInternalCopyStatus {

    var tguid : String? = null

    var statusDEX : String? = null

    var statusEDAV : String? = null

} // .ItemInternalCopyStatus