package util

import org.testng.TestNGException
import java.io.File

class TestFile {
    companion object {
        fun getTestFileFromResources(filename: String): File {
            return File(this::class.java.classLoader.getResource(filename)?.file
                ?: throw TestNGException("Upload test file $filename not found."))
        }
    }
}