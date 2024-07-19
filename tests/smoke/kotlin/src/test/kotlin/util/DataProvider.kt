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
            val manifests = loadManifests(validManifests)
            logger<util.DataProvider>().info("Filtering all valid manifests")
            return filterManifests(manifests)
        }

        @JvmStatic
        @DataProvider(name = "validManifestV1Provider")
        fun validManifestV1Provider(): Array<Array<Map<String, String>>> {
            val validManifests = arrayOf("valid_manifests_v1.json")
            val manifests = loadManifests(validManifests)
            logger<util.DataProvider>().info("Filtering V1 valid manifests")
            return filterManifests(manifests)
        }

        @JvmStatic
        @DataProvider(name = "invalidManifestRequiredFieldsProvider")
        fun invalidManifestRequiredFieldsProvider(): Array<Array<Map<String, String>>> {
            val invalidManifests = arrayOf("invalid_manifests_required_fields.json")
            val manifests = loadManifests(invalidManifests)
            logger<util.DataProvider>().info("Filtering invalid required field manifests")
            return filterManifests(manifests)
        }

        @JvmStatic
        @DataProvider(name = "invalidManifestInvalidValueProvider")
        fun invalidManifestInvalidValueProvider(): Array<Array<Map<String, String>>> {
            val invalidManifests = arrayOf("invalid_manifests_invalid_value.json")
            val manifests = loadManifests(invalidManifests)
            logger<util.DataProvider>().info("Filtering invalid value manifests")
            return filterManifests(manifests)
        }

        private fun loadManifests(manifestFiles: Array<String>): List<Map<String, String>> {
            val manifests = arrayListOf<Map<String, String>>()

            manifestFiles.forEach { manifestFile ->
                val jsonBytes = TestFile.getResourceFile(manifestFile).readBytes()
                val manifestJsons: List<Map<String, String>> = objectMapper.readValue(jsonBytes)
                manifests.addAll(manifestJsons)
            }

            return manifests
        }

        private fun filterManifests(manifests: List<Map<String, String>>): Array<Array<Map<String, String>>> {
            val manifestFilter: String? = System.getProperty("manifestFilter")
            if (manifestFilter.isNullOrEmpty()) {
                return toTypedMatrix(manifests)
            }

            val fields = manifestFilter.split("&")
            val manifestFilters = mutableMapOf<String, List<String>>()

            for (field in fields) {
                val filterTokens = field.split("=")
                if (filterTokens.size != 2) {
                    logger<util.DataProvider>().error("Failed to parse filter for field $field.  Skipping.")
                    continue
                }
                manifestFilters[filterTokens[0]] = filterTokens[1].split(',')
            }

            val filtered = manifests.filter { manifest ->
                manifestFilters.all{ (field, allowedVals) ->
                    manifest[field]?.let { it in allowedVals } ?: false
                }
            }

            logger<util.DataProvider>().info("Found ${filtered.size} manifests out of ${manifests.size}")
            return toTypedMatrix(filtered)
        }

        private fun toTypedMatrix(manifests: List<Map<String, String>>): Array<Array<Map<String, String>>> {
            return manifests.map { arrayOf(it) }.toTypedArray()
        }
    }
}