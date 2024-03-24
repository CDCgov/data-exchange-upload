package util

import org.joda.time.DateTime
import java.io.File
import java.io.FileNotFoundException
import java.util.*
import kotlin.collections.HashMap

class Metadata {
    companion object {
        fun convertPropertiesToMetadataMap(propertiesFilePath: String): HashMap<String, String> {
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

            return metadata
        }

        // Using Calendar due to deprecation of Date.
        fun getFilePrefixByDate(date: DateTime, useCaseDir: String): String {
            // Pad date numbers with 0.
            val month = if (date.monthOfYear < 10) "0${date.monthOfYear}" else "${date.monthOfYear}"
            val day = if (date.dayOfMonth < 10) "0${date.dayOfMonth}" else "${date.dayOfMonth}"
            return "$useCaseDir/${date.year}/$month/$day"
        }
    }
}


