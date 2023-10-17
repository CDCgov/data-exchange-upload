package gov.cdc.ocio.cosmossync.model

import com.google.gson.annotations.SerializedName

// sent by BulkFileUploadFunctionApp to storage queue 
class ItemInternalCopyStatus {

    @SerializedName("Tguid")
    var tguid : String? = null

    @SerializedName("StatusDEX")
    var statusDEX : String? = null

    @SerializedName("StatusEDAV")
    var statusEDAV : String? = null

    override fun toString(): String {
        return "ItemInternalCopyStatus(tguid='$tguid', statusDEX=$statusDEX, statusEDAV=$statusEDAV)"
    } // .toString 

} // .ItemInternalCopyStatus