package util

class Constants {
    companion object {
        val TEST_DESTINATION = "dextesting"
        val TEST_EVENT = "testevent1"
        val BULK_UPLOAD_CONTAINER_NAME = "bulkuploads"
        val EDAV_UPLOAD_CONTAINER_NAME = "upload"
        val ROUTING_UPLOAD_CONTAINER_NAME = "upload"
        val TUS_PREFIX_DIRECTORY_NAME = "tus-prefix"
    }

    class Groups {
        companion object {
            const val PROC_STAT_METADATA_VERIFY_HAPPY_PATH = "proc-stat-metadata-verify-happy-path"
            const val PROC_STAT_UPLOAD_STATUS_HAPPY_PATH = "proc-stat-upload-status-happy-path"
            const val DEX_USE_CASE_DEX_TESTING = "dextesting-testevent1"
        }
    }
}