package util

import org.joda.time.DateTime
import java.io.File
import kotlin.collections.HashMap

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

        // Using Calendar due to deprecation of Date.
        fun getFilePrefixByDate(date: DateTime, useCaseDir: String): String {
            return "$useCaseDir/${date.year}/${date.monthOfYear}/${date.dayOfMonth}"
        }
    }
}