package gov.cdc.ocio.cosmossync.model

class ItemCopyStatus {

    var id : String? = null

    var tguid : String? = null

    var partitionKey: String? = null

    var offset : Long = 0

    var size : Long = 0

    var meta_destination_id: String? = null

    var meta_ext_event: String? = null

    var filename : String? = null

    var metadata: Map<String, Any>? = null

    var start_time_epoch: Long = 0

    var end_time_epoch: Long = 0

    var _ts : Long = 0

    // Item + internal statuses

    var statusDEX : String? = null

    var statusEDAV : String? = null

    override fun toString(): String {
        return "ItemCopyStatus(id='$id', tguid=$tguid, partitionKey=$partitionKey, offset=$offset, size=$size, meta_destination_id=$meta_destination_id, meta_ext_event=$meta_ext_event, filename=$filename, metadata=$metadata, start_time_epoch=$start_time_epoch, end_time_epoch=$end_time_epoch, _ts=$_ts, statusDEX=$statusDEX, statusEDAV=$statusEDAV)"
    } // .toString 
    

} // .ItemCopyStatus