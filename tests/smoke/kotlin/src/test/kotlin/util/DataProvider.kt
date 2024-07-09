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
        fun validManifestAllProvider(): Array<Array<HashMap<String, String>>> {
            val validManifests = arrayOf("valid_manifests_v1.json", "valid_manifests_v2.json")
            return loadAndFilterManifests(validManifests)
        }

        @JvmStatic
        @DataProvider(name = "validManifestV1Provider")
        fun validManifestV1Provider(): Array<Array<HashMap<String, String>>> {
            val validManifests = arrayOf("valid_manifests_v1.json")
            return loadAndFilterManifests(validManifests)
        }

        @JvmStatic
        @DataProvider(name = "invalidManifestRequiredFieldsProvider")
        fun invalidManifestRequiredFieldsProvider(): Array<Array<HashMap<String, String>>> {
            val invalidManifests = arrayOf("invalid_manifests_required_fields.json")
            return loadAndFilterManifests(invalidManifests)
        }

        @JvmStatic
        @DataProvider(name = "invalidManifestInvalidValueProvider")
        fun invalidManifestInvalidValueProvider(): Array<Array<HashMap<String, String>>> {
            val invalidManifests = arrayOf("invalid_manifests_invalid_value.json")
            return loadAndFilterManifests(invalidManifests)
        }

        private fun loadAndFilterManifests(manifestFiles: Array<String>): Array<Array<HashMap<String, String>>> {
            val useCases: String? = System.getProperty("useCases")
            val useCaseFilters: List<Map<String, String>> = useCases?.split(";")?.map { useCase ->
                useCase.split(",").associate {
                    val (key, value) = it.split(":")
                    key to value
                }
            } ?: listOf()

            val manifests = arrayListOf<HashMap<String, String>>()

            manifestFiles.forEach { manifestFile ->
                val jsonBytes = TestFile.getResourceFile(manifestFile).readBytes()
                val manifestJsons: List<HashMap<String, String>> = objectMapper.readValue(jsonBytes)
                useCaseFilters.forEach { manifestFilter ->
                    val filtered = filterByUseCases(manifestJsons, manifestFilter)
                    manifests.addAll(filtered)
                }
            }

            return manifests.map { arrayOf(it) }.toTypedArray()
        }

        private fun filterByUseCases(
            manifests: List<HashMap<String, String>>,
            manifestFilter: Map<String, String>
        ): List<HashMap<String, String>> {
            return if (manifestFilter.isNotEmpty()) {
                manifests.filter { manifest ->
                    manifestFilter.all { (key, value) ->
                        manifest[key] == value
                    }
                }
            } else {
                manifests
            }
        }
    }
}