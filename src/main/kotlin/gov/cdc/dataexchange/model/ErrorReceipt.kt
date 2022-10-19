package gov.cdc.dataexchange.model

import com.fasterxml.jackson.annotation.JsonInclude
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter

@JsonInclude(JsonInclude.Include.NON_NULL)
class ErrorReceipt() {
    private var code: ErrorCodes? = null
    private val timestamp: String = ZonedDateTime.now().format(DateTimeFormatter.ISO_INSTANT)
    private var description: String? = null
    private var status = 0
    private var path: String? = null
    private var exception: String? = null
    var details: Array<Any> = arrayOf()

    constructor(code: ErrorCodes?, description: String?, status: Int, path: String?, exception: String?) : this() {
        this.code = code
        this.description = description
        this.status = status
        this.path = path
        this.exception = exception
    }
}