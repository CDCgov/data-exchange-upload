package gov.cdc.dataexchange.controller

import com.google.gson.Gson
import gov.cdc.dataexchange.model.ErrorCodes
import gov.cdc.dataexchange.model.ErrorReceipt
import gov.cdc.dataexchange.model.MissingFileException
import org.springframework.http.HttpHeaders
import org.springframework.http.HttpStatus
import org.springframework.http.MediaType
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.ExceptionHandler
import org.springframework.web.bind.annotation.RestControllerAdvice
import javax.servlet.http.HttpServletRequest


@RestControllerAdvice
class ExceptionHandlers {

    @ExceptionHandler(MissingFileException::class)
    fun handleMissingFileException(request: HttpServletRequest,
                                   ex: MissingFileException): ResponseEntity<*>? {
        val headers = HttpHeaders()
        headers.contentType = MediaType.APPLICATION_JSON
        val error = ErrorReceipt(
            ErrorCodes.BAD_REQUEST,
            "Invalid Upload",
            HttpStatus.BAD_REQUEST.value(),
            request.requestURL.toString(),
            MissingFileException::class.java.name
        )
        return ResponseEntity.badRequest().headers(headers).body<String>(Gson().toJson(error))
    }

}