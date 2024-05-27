package util

import org.testng.TestNGException
import java.io.File

class TestFile {
    companion object {
        fun getResourceFile(filename: String): File {
            return File(
                this::class.java.classLoader.getResource(filename)?.file
                    ?: throw TestNGException("file $filename not found.")
            )
        }
    }
}
