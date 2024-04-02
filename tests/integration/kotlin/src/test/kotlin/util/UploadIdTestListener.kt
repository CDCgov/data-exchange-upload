package util

import org.testng.ITestResult
import org.testng.TestListenerAdapter

class UploadIdTestListener: TestListenerAdapter() {
    override fun onTestFailure(tr: ITestResult?) {
        super.onTestFailure(tr)

        tr?.testContext?.getAttribute("uploadId")?.toString()?.let {
            println("Test ${tr.testClass.name}/${tr.method.methodName} failed for Upload ID: $it")
        }
    }
}