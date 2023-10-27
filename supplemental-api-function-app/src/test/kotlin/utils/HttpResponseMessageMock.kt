package utils

import com.microsoft.azure.functions.HttpResponseMessage
import com.microsoft.azure.functions.HttpStatus
import com.microsoft.azure.functions.HttpStatusType


class HttpResponseMessageMock(private val httpStatus: HttpStatusType, headers: Map<String, String>?, body: Any) :
    HttpResponseMessage {
    private val httpStatusCode: Int
    private val body: Any
    private val headers: Map<String, String>?

    init {
        httpStatusCode = httpStatus.value()
        this.headers = headers
        this.body = body
    }

    override fun getStatus(): HttpStatusType {
        return httpStatus
    }

    override fun getStatusCode(): Int {
        return httpStatusCode
    }

    override fun getHeader(key: String): String {
        return headers!![key]!!
    }

    override fun getBody(): Any {
        return body
    }

    class HttpResponseMessageBuilderMock : HttpResponseMessage.Builder {
        private var body: Any? = null
        private var httpStatusCode = 0
        private val headers: MutableMap<String, String> = mutableMapOf()
        private var httpStatus: HttpStatusType? = null
        fun status(status: HttpStatus): HttpResponseMessage.Builder {
            httpStatusCode = status.value()
            httpStatus = status
            return this
        }

        override fun status(httpStatusType: HttpStatusType): HttpResponseMessage.Builder {
            httpStatusCode = httpStatusType.value()
            httpStatus = httpStatusType
            return this
        }

        override fun header(key: String, value: String): HttpResponseMessage.Builder {
            headers!![key] = value
            return this
        }

        override fun body(body: Any): HttpResponseMessage.Builder {
            this.body = body
            return this
        }

        override fun build(): HttpResponseMessage {
            return HttpResponseMessageMock(httpStatus!!, headers, body!!)
        }
    }
}