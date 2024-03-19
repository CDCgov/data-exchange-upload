package util

import java.io.File
import java.io.FileNotFoundException
import java.io.FileReader
import java.util.*
import kotlin.collections.HashMap

class Metadata {
    companion object {
//        fun generateRequiredMetadataForFile(file: File): HashMap<String, String> {
//            return hashMapOf(
//                "filename" to file.name,
//                "meta_destination_id" to Constants.TEST_DESTINATION,
//                "meta_ext_event" to Constants.TEST_EVENT,
//                "meta_ext_source" to "INTEGRATION-TEST"
//            )
//
//        }

        // this method is for reading the values from properties file and creating key value pair for metadata
        fun generateRequiredMetadataForFile(file: File, propertiesFilePath: String): HashMap<String, String> {
            val metadata = HashMap<String, String>()
            val properties = Properties()
            val inputStream = this::class.java.classLoader.getResourceAsStream(propertiesFilePath)
                    ?: throw FileNotFoundException("Property file '$propertiesFilePath' not found in the classpath")

            inputStream.use { stream ->
                properties.load(stream)
            }

            properties.forEach { key, value ->
                metadata[key as String] = value as String
            }
            println("metadata: "+ metadata)

            return metadata
        }

    }
}


