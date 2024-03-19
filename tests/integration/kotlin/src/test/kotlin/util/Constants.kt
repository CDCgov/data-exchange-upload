package util

import java.util.*

class Constants {
    companion object {
        val TEST_DESTINATION = "dextesting"
        val TEST_EVENT = "testevent1"
          val properties = Properties()

        init {
            val configFile = System.getProperty("CONFIG_FILE", "")
            println("configFile"+ configFile)
            val inputStream = javaClass.classLoader.getResourceAsStream("properties/"+configFile)

            properties.load(inputStream)
        }

//        val TEST_DESTINATION: String = properties.getProperty("TEST_DESTINATION", "")
//        val TEST_EVENT: String = properties.getProperty("TEST_EVENT", "")

        const val BULK_UPLOAD_CONTAINER_NAME = "bulkuploads"
        const val EDAV_UPLOAD_CONTAINER_NAME = "upload"
        const val ROUTING_UPLOAD_CONTAINER_NAME = "routeingress"
        const val TUS_PREFIX_DIRECTORY_NAME = "tus-prefix"
    }


    class Groups {
        companion object {
            const val PROC_STAT_METADATA_VERIFY_HAPPY_PATH = "proc-stat-metadata-verify-happy-path"
            const val PROC_STAT_UPLOAD_FILE_HAPPY_PATH = "proc-stat-upload-file-happy-path"
            const val PROC_STAT_UPLOAD_STATUS_DEX_FILE_COPY_HAPPY_PATH="proc-stat-upload-status-dex-file-copy-happy-path"
            const val DEX_USE_CASE_DEX_TESTING = "dextesting-testevent1"
        }

    }
}
