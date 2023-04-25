package gov.cdc.ocio.supplementalapi.model

class UploadStatus {

    var status: String? = null

    var percent_complete: Float = 0F

    var file_name: String? = null

    var file_size_bytes: Long = 0

    var bytes_uploaded: Long = 0

    var tus_upload_id: String? = null

    var time_uploading_sec: Long = 0

    var metadata: Map<String, Any>? = null

    var timestamp: String? = null

    companion object {

        fun createFromItem(item: Item): UploadStatus {
            var percentComplete = 0F
            if (item.size > 0)
                percentComplete = item.offset.toFloat() / item.size * 100

            val statusMessage = if (item.offset < item.size) "Uploading" else "Complete"
            return UploadStatus().apply {
                status = statusMessage
                tus_upload_id = item.tguid
                file_name = item.filename
                file_size_bytes = item.size
                bytes_uploaded = item.offset
                percent_complete = percentComplete
                time_uploading_sec = item.end_time_epoch - item.start_time_epoch
                metadata = item.metadata
                timestamp = item.getTimestamp()
            }
        }
    }
}