package util

import model.CopyConfig
import org.joda.time.DateTime
import org.joda.time.DateTimeZone
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

        fun getFilePrefixByDate(date: DateTime, useCaseDir: String): String {
            // Pad date numbers with 0.
            val month = if (date.monthOfYear < 10) "0${date.monthOfYear}" else "${date.monthOfYear}"
            val day = if (date.dayOfMonth < 10) "0${date.dayOfMonth}" else "${date.dayOfMonth}"
            return "$useCaseDir/${date.year}/$month/$day"
        }

        fun getFilePrefix(copyConfig: CopyConfig): String {
            var prefix = ""

            if (copyConfig.folderStructure == "date_YYYY_MM_DD") {
                prefix = "${getFilePrefixByDate(DateTime(DateTimeZone.UTC))}/"
            }

            return prefix
        }

        fun getFilePrefix(copyConfig: CopyConfig, manifest: Map<String, String>): String {
            var prefix = "${getUseCaseFromManifest(manifest)}/"

            if (copyConfig.folderStructure == "date_YYYY_MM_DD") {
                prefix = "$prefix${getFilePrefixByDate(DateTime(DateTimeZone.UTC))}/"
            }

            return prefix
        }

        fun getFilePrefixByDate(date: DateTime, manifest: HashMap<String, String>): String {
            // Pad date numbers with 0.
            val month = if (date.monthOfYear < 10) "0${date.monthOfYear}" else "${date.monthOfYear}"
            val day = if (date.dayOfMonth < 10) "0${date.dayOfMonth}" else "${date.dayOfMonth}"
            return "${getUseCaseFromManifest(manifest)}/${date.year}/$month/$day"
        }

        // Using Calendar due to deprecation of Date.
        fun getFilePrefixByDate(date: DateTime): String {
            // Pad date numbers with 0.
            val month = if (date.monthOfYear < 10) "0${date.monthOfYear}" else "${date.monthOfYear}"
            val day = if (date.dayOfMonth < 10) "0${date.dayOfMonth}" else "${date.dayOfMonth}"
            return "${date.year}/$month/$day"
        }

        fun getUseCaseFromManifest(manifest: Map<String, String>): String {
            return if (manifest.containsKey("version")) {
                when (manifest["version"]) {
                    "2.0" -> "${manifest["data_stream_id"]}-${manifest["data_stream_route"]}"
                    else -> "${manifest["meta_destination_id"]}-${manifest["meta_ext_event"]}"
                }
            } else {
                "${manifest["meta_destination_id"]}-${manifest["meta_ext_event"]}"
            }
        }

        fun getFilename(manifest: Map<String, String>): String {
            val filenameKeys = arrayOf("received_filename", "meta_ext_filename", "filename", "original_filename")

            return manifest.entries.first { filenameKeys.contains(it.key) }.value
        }

        private fun getMetadataPath(version: String, useCase: String, manifest: String): String {
            return "properties/$version/$useCase/$manifest"
        }
    }
}