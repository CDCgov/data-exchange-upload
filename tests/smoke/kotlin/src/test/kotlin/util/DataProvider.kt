package util

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
            return arrayOf(
                arrayOf(
                    hashMapOf(
                        "meta_destination_id" to "dextesting",
                        "meta_ext_event" to "testevent1",
                        "filename" to "dex-smoke-test"
                    ),
                ),
                arrayOf(
                    hashMapOf(
                        "meta_destination_id" to "aims-celr",
                        "meta_ext_event" to "csv",
                        "meta_ext_objectkey" to "test-obj-key",
                        "meta_ext_filename" to "dex-smoke-test",
                        "meta_ext_file_timestamp" to "test-timestamp",
                        "meta_username" to "test-automation",
                        "meta_ext_filestatus" to "test-file-status",
                        "meta_ext_uploadid" to "test-upload-id",
                        "meta_ext_source" to "test-src",
                        "meta_organization" to "test-org"
                    )
                )
            )
        }
    }
}