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
            val useCases: String? = System.getProperty("useCases")
            val validManifests = arrayOf("valid_manifests_v1.json", "valid_manifests_v2.json")
            val manifests = arrayListOf<HashMap<String, String>>()
            val manifestFilter: HashMap<String, String> = useCases?.split(",")?.associate {
                val (key, value) = it.split(":")
                key to value
            }?.let { HashMap(it) } ?: HashMap()

            validManifests.forEach {
                val jsonBytes = TestFile.getResourceFile(it).readBytes()
                val manifestJsons: List<HashMap<String, String>> = ObjectMapper().readValue(jsonBytes)
                val filtered = filterByUseCases(manifestJsons, manifestFilter)
                manifests.addAll(filtered)
            }

            return manifests.map { arrayOf(it) }.toTypedArray()
        }

        @DataProvider(name = "validManifestV1Provider")
        @JvmStatic
        fun validManifestV1Provider(): Array<Array<HashMap<String, String>>> {
            val useCases: String? = System.getProperty("useCases")
            val validManifests = arrayOf("valid_manifests_v1.json")
            val manifests = arrayListOf<HashMap<String, String>>()
            val manifestFilter: HashMap<String, String> = useCases?.split(",")?.associate {
                val (key, value) = it.split(":")
                key to value
            }?.let { HashMap(it) } ?: HashMap()

            validManifests.forEach {
                val jsonBytes = TestFile.getResourceFile(it).readBytes()
                val manifestJsons: List<HashMap<String, String>> = ObjectMapper().readValue(jsonBytes)
                val filtered = filterByUseCases(manifestJsons, manifestFilter)
                manifests.addAll(filtered)
            }

            return manifests.map { arrayOf(it) }.toTypedArray()
        }

        @DataProvider(name = "invalidManifestRequiredFieldsProvider")
        @JvmStatic
        fun invalidManifestRequiredFieldsProvider(): Array<Array<HashMap<String, String>>> {
            val useCases: String? = System.getProperty("useCases")
            val validManifests = arrayOf("invalid_manifests_required_fields.json")
            val manifests = arrayListOf<HashMap<String, String>>()
            val manifestFilter: HashMap<String, String> = useCases?.split(",")?.associate {
                val (key, value) = it.split(":")
                key to value
            }?.let { HashMap(it) } ?: HashMap()

            validManifests.forEach {
                val jsonBytes = TestFile.getResourceFile(it).readBytes()
                val manifestJsons: List<HashMap<String, String>> = ObjectMapper().readValue(jsonBytes)
                val filtered = filterByUseCases(manifestJsons, manifestFilter)
                manifests.addAll(filtered)
            }

            return manifests.map { arrayOf(it) }.toTypedArray()
        }

        @DataProvider(name = "invalidManifestInvalidValueProvider")
        @JvmStatic
        fun invalidManifestInvalidValueProvider(): Array<Array<HashMap<String, String>>> {
            val useCases: String? = System.getProperty("useCases")
            val validManifests = arrayOf("invalid_manifests_invalid_value.json")
            val manifests = arrayListOf<HashMap<String, String>>()
            val manifestFilter: HashMap<String, String> = useCases?.split(",")?.associate {
                val (key, value) = it.split(":")
                key to value
            }?.let { HashMap(it) } ?: HashMap()

            validManifests.forEach {
                val jsonBytes = TestFile.getResourceFile(it).readBytes()
                val manifestJsons: List<HashMap<String, String>> = ObjectMapper().readValue(jsonBytes)
                val filtered = filterByUseCases(manifestJsons, manifestFilter)
                manifests.addAll(filtered)
            }

            return manifests.map { arrayOf(it) }.toTypedArray()
        }

        private fun filterByUseCases(
            manifests: List<HashMap<String, String>>,
            manifestFilter: HashMap<String, String>
        ): List<HashMap<String, String>> {
            return if (manifestFilter.isNotEmpty()) {
                manifests.filter { m ->
                    manifestFilter.all { (key, value) ->
                        m[key] == value
                    }
                }
            } else {
                manifests
            }
        }
    }
}