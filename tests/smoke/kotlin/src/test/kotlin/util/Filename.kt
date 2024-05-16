package util

import model.CopyConfig

class Filename {

    companion object {
        fun getFilenameSuffix(copyConfig: CopyConfig, uploadId: String): String {
            return if (copyConfig.filenameSuffix == "upload_id") "_$uploadId" else ""
        }
    }
}