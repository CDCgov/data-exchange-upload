package util

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import org.testng.annotations.DataProvider

class DataProvider {
    companion object {
        private val objectMapper = ObjectMapper()

        @JvmStatic
        @DataProvider(name = "versionProvider")
        fun versionProvider(): Array<Array<String>> {
            return arrayOf(
                arrayOf("v1"),
                arrayOf("v2")
            )
        }

        @JvmStatic
        @DataProvider(name = "validManifestAllProvider")
        fun validManifestAllProvider(): Array<Array<Map<String, String>>> {
            val validManifests = arrayOf("valid_manifests_v1.json", "valid_manifests_v2.json")
            return loadAndFilterManifests(validManifests)
        }

        @JvmStatic
        @DataProvider(name = "validManifestV1Provider")
        fun validManifestV1Provider(): Array<Array<Map<String, String>>> {
            val validManifests = arrayOf("valid_manifests_v1.json")
            return loadAndFilterManifests(validManifests)
        }

        @JvmStatic
        @DataProvider(name = "invalidManifestRequiredFieldsProvider")
        fun invalidManifestRequiredFieldsProvider(): Array<Array<Map<String, String>>> {
            val invalidManifests = arrayOf("invalid_manifests_required_fields.json")
            return loadAndFilterManifests(invalidManifests)
        }

        @JvmStatic
        @DataProvider(name = "invalidManifestInvalidValueProvider")
        fun invalidManifestInvalidValueProvider(): Array<Array<Map<String, String>>> {
            val invalidManifests = arrayOf("invalid_manifests_invalid_value.json")
            return loadAndFilterManifests(invalidManifests)
        }

        private fun loadAndFilterManifests(manifestFiles: Array<String>): Array<Array<Map<String, String>>> {
            val manifestFilter: String? = System.getProperty("manifestFilter")
            val fields = manifestFilter?.split(";")

            val manifestFilters = mutableMapOf<String, String>()

            if (fields != null) {
                for (field in fields) {
                    val keyValue = field.split("=")
                    if (keyValue.size == 2) {
                        manifestFilters[keyValue[0]] = keyValue[1]
                    }
                }
            }

            val manifests = arrayListOf<Map<String, String>>()
            var totalManifests = 0

            manifestFiles.forEach { manifestFile ->
                val jsonBytes = TestFile.getResourceFile(manifestFile).readBytes()
                val manifestJsons: List<Map<String, String>> = objectMapper.readValue(jsonBytes)
                totalManifests += manifestJsons.size

                if (manifestFilters.isNotEmpty()) {
                    val filtered = filterManifestJsons(manifestJsons, manifestFilters)
                    println("Filtered manifests: $filtered")
                    manifests.addAll(filtered)
                } else {
                    manifests.addAll(manifestJsons)
                }
            }
            println("Total number of manifests: $totalManifests, Number of filtered manifests: ${manifests.size}")
            println("Final Manifest: $manifests")
            return manifests.map { arrayOf(it) }.toTypedArray()
        }

        private fun parseFilterValues(filter: String): List<String> {
            try {
                val values = filter.split(",")
                if (values.size != 2) {
                    if (values.size == 1) {
                        return values
                    } else {
                        throw IllegalArgumentException("Filter values must contain exactly two elements.")
                    }
                }
                return values
            } catch (e: Exception) {
                throw RuntimeException("An error occurred while parsing filter values.", e)
            }
        }

        private fun filterManifestJsons(
            manifestJsons: List<Map<String, String>>,
            manifestFilters: Map<String, String>
        ): List<Map<String, String>> {
            return manifestJsons.filter { json ->
                manifestFilters.all { (key, value) ->
                    val filterValues = parseFilterValues(value)
                    json[key]?.let { it in filterValues } ?: false
                }
            }
        }
    }
}