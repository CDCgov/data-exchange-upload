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

        @DataProvider(name = "validManifestProvider")
        @JvmStatic
        fun validManifestProvider(): Array<Array<HashMap<String, String>>> {
            val useCases: List<String> = System.getProperty("useCases")?.split(",") ?: arrayListOf()
            val jsonBytes = TestFile.getResourceFile("valid_manifests.json").readBytes()
            val manifests: Array<HashMap<String, String>> = ObjectMapper().readValue(jsonBytes)

            // Filter by use cases.
            val filtered: Array<HashMap<String, String>> = if (useCases.isNotEmpty()) manifests.filter {
                val useCase = if (it.containsKey("version") && it["version"] == "2.0")
                    "${it["data_stream_id"]}-${it["data_stream_route"]}"
                else "${it["meta_destination_id"]}-${it["meta_ext_event"]}"

                useCases.contains(useCase)
            }.toTypedArray() else manifests

            return filtered .map { arrayOf(it) }.toTypedArray()
        }
    }
}