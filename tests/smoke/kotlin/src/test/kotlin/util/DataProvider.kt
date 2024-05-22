package util

import org.testng.annotations.DataProvider

class DataProvider {
    companion object {
        @JvmStatic
        @DataProvider(name = "versionProvider")
        fun versionProvider(): Array<Array<String>> {
            return arrayOf(
                arrayOf("V1"),
                arrayOf("V2")
            )
        }
    }
}