package util

class Constants {
    companion object {
        const val TEST_DESTINATION = "dextesting"
        const val TEST_EVENT = "testevent1"
        const val BULK_UPLOAD_CONTAINER_NAME = "bulkuploads"
        const val UPLOAD_CONFIG_CONTAINER_NAME = "upload-configs"
        const val EDAV_UPLOAD_CONTAINER_NAME = "upload"
        const val ROUTING_UPLOAD_CONTAINER_NAME = "routeingress"
        const val TUS_PREFIX_DIRECTORY_NAME = "tus-prefix"
    }

    class Groups {
        companion object {
            const val METADATA_VERIFY = "metadata-verify"
            const val PROC_STAT = "proc-stat"
            const val PROC_STAT_TRACE = "proc-stat-trace"
            const val PROC_STAT_REPORT = "proc-stat-report"
            const val FILE_COPY = "file-copy"
        }
    }
}