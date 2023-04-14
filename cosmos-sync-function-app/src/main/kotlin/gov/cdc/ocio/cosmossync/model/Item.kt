package gov.cdc.ocio.cosmossync.model

class Item {

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
}