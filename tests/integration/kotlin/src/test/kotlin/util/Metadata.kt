package util

import java.io.File

class Metadata {
    companion object {
        fun generateRequiredMetadataForFile(file: File): HashMap<String, String> {
            return hashMapOf(
                "filename" to file.name,
                "meta_destination_id" to Constants.TEST_DESTINATION,
                "meta_ext_event" to Constants.TEST_EVENT,
                "meta_ext_source" to "INTEGRATION-TEST"
            )
        }
    }
}