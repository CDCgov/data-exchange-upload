package gov.cdc.ocio.supplementalapi.model

import java.text.SimpleDateFormat
import java.util.*

class Item {

    var tguid : String? = null

    var offset : Long = 0

    var size : Long = 0

    var filename : String? = null

    var metadata : Map<String, Any>? = null

    var start_time_epoch: Long = 0

    var end_time_epoch: Long = 0

    var _ts : Long = 0

    fun getTimestamp(): String {
        val date = Date(_ts * 1000) // convert to milliseconds
        val sdf = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss'Z'", Locale.ENGLISH)
        sdf.timeZone = TimeZone.getTimeZone("UTC")
        return sdf.format(date)
    }

}