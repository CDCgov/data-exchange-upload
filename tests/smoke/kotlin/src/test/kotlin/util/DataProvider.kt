package util

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import org.testng.annotations.DataProvider

class DataProvider {
    companion object {
        @JvmStatic
        @DataProvider(name = "versionProvider")
        fun versionProvider(): Array<Array<String>> {
            return arrayOf(
                arrayOf("v1"),
                arrayOf("v2")
            )
        }

        @DataProvider(name = "validManifestAllProvider")
        @JvmStatic
        fun validManifestAllProvider(): Array<Array<HashMap<String, String>>> {
            val useCases: List<String> = System.getProperty("useCases")?.split(",") ?: arrayListOf()
            val validManifests = arrayOf("valid_manifests_v1.json", "valid_manifests_v2.json")
            val manifests = arrayListOf<HashMap<String, String>>()

            val criteriaString: String? = System.getProperty("criteria")
            val criteria: Map<String, String> = criteriaString?.split(",")?.associate {
                val (key, value) = it.split(":")
                key to value
            } ?: emptyMap()

            validManifests.forEach {
                val jsonBytes = TestFile.getResourceFile(it).readBytes()
                val manifestJsons: List<HashMap<String, String>> = ObjectMapper().readValue(jsonBytes)
                val filtered = filterByUseCases(useCases, manifestJsons, criteria)
                manifests.addAll(filtered)
            }

            return manifests.map { arrayOf(it) }.toTypedArray()
        }

        @DataProvider(name = "validManifestV1Provider")
        @JvmStatic
        fun validManifestV1Provider(): Array<Array<HashMap<String, String>>> {
            val useCases: List<String> = System.getProperty("useCases")?.split(",") ?: arrayListOf()

            val criteriaString: String? = System.getProperty("criteria")
            val criteria: Map<String, String> = criteriaString?.split(",")?.associate {
                val (key, value) = it.split(":")
                key to value
            } ?: emptyMap()

            val jsonBytes = TestFile.getResourceFile("valid_manifests_v1.json").readBytes()
            val manifestJsons: List<HashMap<String, String>> = ObjectMapper().readValue(jsonBytes)
            val manifests = filterByUseCases(useCases, manifestJsons, criteria)

            return manifests.map { arrayOf(it) }.toTypedArray()
        }

        @DataProvider(name = "invalidManifestRequiredFieldsProvider")
        @JvmStatic
        fun invalidManifestRequiredFieldsProvider(): Array<Array<HashMap<String, String>>> {
            val useCases: List<String> = System.getProperty("useCases")?.split(",") ?: arrayListOf()

            val criteriaString: String? = System.getProperty("criteria")
            val criteria: Map<String, String> = criteriaString?.split(",")?.associate {
                val (key, value) = it.split(":")
                key to value
            } ?: emptyMap()

            val jsonBytes = TestFile.getResourceFile("invalid_manifests_required_fields.json").readBytes()
            val manifestJsons: List<HashMap<String, String>> = ObjectMapper().readValue(jsonBytes)
            val manifests = filterByUseCases(useCases, manifestJsons, criteria)

            return manifests.map { arrayOf(it) }.toTypedArray()
        }

        @DataProvider(name = "invalidManifestInvalidValueProvider")
        @JvmStatic
        fun invalidManifestInvalidValueProvider(): Array<Array<HashMap<String, String>>> {
            val useCases: List<String> = System.getProperty("useCases")?.split(",") ?: arrayListOf()

            val criteriaString: String? = System.getProperty("criteria")
            val criteria: Map<String, String> = criteriaString?.split(",")?.associate {
                val (key, value) = it.split(":")
                key to value
            } ?: emptyMap()

            val jsonBytes = TestFile.getResourceFile("invalid_manifests_invalid_value.json").readBytes()
            val manifestJsons: List<HashMap<String, String>> = ObjectMapper().readValue(jsonBytes)
            val manifests = filterByUseCases(useCases, manifestJsons, criteria)

            return manifests.map { arrayOf(it) }.toTypedArray()
        }

        private fun filterByUseCases(
            useCases: List<String>,
            manifests: List<HashMap<String, String>>,
            criteria: Map<String, String>
        ): List<HashMap<String, String>> {
            return if (useCases.isNotEmpty()) {
                manifests.filter { m ->
                    val useCase = if (m.containsKey("version") && m["version"] == "2.0")
                        "${m["data_stream_id"]}-${m["data_stream_route"]}"
                    else "${m["meta_destination_id"]}-${m["meta_ext_event"]}"

                    useCases.contains(useCase) && criteria.all { (key, value) ->
                        m[key] == value
                    }
                }
            } else {
                manifests
            }
        }
    }
}