package util

import com.fasterxml.jackson.annotation.JsonProperty
import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.readValue
import org.testng.annotations.DataProvider

class DataProvider {
    companion object {
        private val objectMapper = ObjectMapper()

        @JvmStatic
        @DataProvider(name = "validManifestProvider")
        fun validManifestAllProvider(): Array<Array<TestCase>> {
            var validManifests = arrayOf("valid_manifests.json")
            val cases = loadTestCases(validManifests)
            logger<util.DataProvider>().info("Filtering all valid manifests")
            return filterCases(cases)
        }

        @JvmStatic
        @DataProvider(name = "invalidManifestRequiredFieldsProvider")
        fun invalidManifestRequiredFieldsProvider(): Array<Array<TestCase>> {
            val invalidManifests = arrayOf("invalid_manifests_required_fields.json")
            val manifests = loadTestCases(invalidManifests)
            logger<util.DataProvider>().info("Filtering invalid required field manifests")
            return filterCases(manifests)
        }

        @JvmStatic
        @DataProvider(name = "invalidManifestInvalidValueProvider")
        fun invalidManifestInvalidValueProvider(): Array<Array<TestCase>> {
            val invalidManifests = arrayOf("invalid_manifests_invalid_value.json")
            val manifests = loadTestCases(invalidManifests)
            logger<util.DataProvider>().info("Filtering invalid value manifests")
            return filterCases(manifests)
        }

        private fun loadTestCases(caseFiles: Array<String>): List<TestCase> {
            val cases = arrayListOf<TestCase>()

            caseFiles.forEach { caseFile ->
                val jsonBytes = TestFile.getResourceFile(caseFile).readBytes()
                val caseJson: List<TestCase> = objectMapper.readValue(jsonBytes)
                cases.addAll(caseJson)
            }

            return cases
        }

        private fun filterCases(cases: List<TestCase>): Array<Array<TestCase>> {
            val manifestFilter: String? = System.getProperty("manifestFilter")
            if (manifestFilter.isNullOrEmpty()) {
                return toTypedMatrix(cases)
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

            val filtered = cases.filter { case ->
                manifestFilters.all{ (field, allowedVals) ->
                    case.manifest[field]?.let { it in allowedVals } ?: false
                }
            }

            logger<util.DataProvider>().info("Found ${filtered.size} manifests out of ${cases.size}")
            return toTypedMatrix(filtered)
        }

        private inline fun <reified T> toTypedMatrix(items: List<T>): Array<Array<T>> {
            return items.map { arrayOf(it) }.toTypedArray()
        }
    }
}

data class TestCase(
    @JsonProperty("manifest")val manifest: Map<String, String>,
    @JsonProperty("delivery_targets") val deliveryTargets: List<Target>?
)

data class Target(
    @JsonProperty("name") val name: String,
    @JsonProperty("path_template") val pathTemplate: Map<Environment, String>
)
