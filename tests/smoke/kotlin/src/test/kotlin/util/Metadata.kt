package util

import com.fasterxml.jackson.databind.ObjectMapper
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

        fun getFilePrefixByDate(date: DateTime, useCaseDir: String): String {
            // Pad date numbers with 0.
            val month = if (date.monthOfYear < 10) "0${date.monthOfYear}" else "${date.monthOfYear}"
            val day = if (date.dayOfMonth < 10) "0${date.dayOfMonth}" else "${date.dayOfMonth}"
            return "$useCaseDir/${date.year}/$month/$day"
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

        fun getSenderManifest(version: String, useCase: String, manifest: String): HashMap<String, String> {
            val path = getMetadataPath(version, useCase, manifest)
            return convertPropertiesToMetadataMap(path)
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

        fun getFilename(manifest: HashMap<String, String>): String {
            val filenameKeys = arrayOf("received_filename", "meta_ext_filename", "filename", "original_filename")

            return manifest.entries.first { filenameKeys.contains(it.key) }.value
        }

        private fun getMetadataPath(version: String, useCase: String, manifest: String): String {
            return "properties/$version/$useCase/$manifest"
        }

        fun readMetadataFromJsonFile(useCase: String): HashMap<String, String> {
            val objectMapper = ObjectMapper()
            val metadataMap = HashMap<String, String>()
            val jsonFiles = listOf("valid_manifests_v1.json", "valid_manifests_v2.json")

            jsonFiles.forEach { jsonFile ->
                val inputStream = this::class.java.classLoader.getResourceAsStream(jsonFile)
                    ?: throw FileNotFoundException("JSON file '$jsonFile' not found in the classpath")

                val jsonNodes = objectMapper.readTree(inputStream)
                jsonNodes.forEach { node ->
                    val nodeMap = node.fields().asSequence().map { it.key to it.value.asText() }.toMap()
                    if (getUseCaseFromManifest(nodeMap) == useCase) {
                        node.fields().forEachRemaining { entry ->
                            metadataMap[entry.key] = entry.value.asText()
                        }
                    }
                }
            }
            return metadataMap
        }
    }
}