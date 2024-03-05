package util

class Constants {
    companion object {
        const val TEST_DESTINATION = "dextesting"
        const val TEST_EVENT = "testevent1"
        const val BULK_UPLOAD_CONTAINER_NAME = "bulkuploads"
        const val EDAV_UPLOAD_CONTAINER_NAME = "upload"
        const val ROUTING_UPLOAD_CONTAINER_NAME = "routeingress"
        const val TUS_PREFIX_DIRECTORY_NAME = "tus-prefix"
    }

    class Groups {
        companion object {
            const val PROC_STAT_METADATA_VERIFY_HAPPY_PATH = "proc-stat-metadata-verify-happy-path"
            const val PROC_STAT_UPLOAD_STATUS_HAPPY_PATH = "proc-stat-upload-status-happy-path"
            const val DEX_USE_CASE_DEX_TESTING = "dextesting-testevent1"
        }
    }
}