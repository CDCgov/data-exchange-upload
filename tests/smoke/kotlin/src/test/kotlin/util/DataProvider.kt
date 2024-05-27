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
            val jsonBytes = TestFile.getResourceFile("valid_manifests.json").readBytes()
            val manifests: Array<HashMap<String, String>> = ObjectMapper().readValue(jsonBytes)

            return manifests.map { arrayOf(it) }.toTypedArray()
        }
    }
}