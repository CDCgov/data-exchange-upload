package gov.cdc.dataexchange.controller

import com.google.gson.Gson
import gov.cdc.dataexchange.azure.BlobProxy
import gov.cdc.dataexchange.model.ErrorCodes
import gov.cdc.dataexchange.model.ErrorReceipt
import gov.cdc.dataexchange.model.MissingFileException
import org.slf4j.LoggerFactory
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.http.HttpStatus
import org.springframework.http.MediaType
import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.*
import org.springframework.web.multipart.MultipartFile
import javax.servlet.http.HttpServletRequest


@RestController
@RequestMapping("/upload")
class UploadController {

    @Autowired
    private var proxy: BlobProxy? = null

    private val log = LoggerFactory.getLogger(UploadController::class.java)

    @PostMapping(path = ["/file"], produces = [MediaType.APPLICATION_JSON_VALUE])
    @Throws(MissingFileException::class)
    fun processFile(file: MultipartFile?,
                    parameters: MutableMap<String?, String?>,
                    request: HttpServletRequest
    ): ResponseEntity<*>? {
        if (file == null) throw MissingFileException()

        parameters["meta_ext_filename"] = file.originalFilename
        log.info("==> The file '{}' with size:'{}kB' is being processed", file.originalFilename, file.size / 1024)
        log.info("params: {}", parameters)
        return try {
            val start = System.currentTimeMillis()
            val receipt = proxy?.uploadFile(file, parameters)
            val end = System.currentTimeMillis()
            log.info(
                "Accepted message tguid: {}, etag: {}. elapsed time: {} ms",
                receipt?.tguid, receipt?.etag, (end - start) / 1e6
            )
            ResponseEntity.status(HttpStatus.ACCEPTED).body(Gson().toJson(receipt))
        } catch (e: Exception) {
            log.error("exception: " + e.message)
            val error = ErrorReceipt(
                ErrorCodes.INTERNAL_SERVER_ERROR,
                e.message,
                HttpStatus.INTERNAL_SERVER_ERROR.value(),
                request.requestURL.toString(),
                e.javaClass.name
            )
            ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).body<Any>(error)
        }
    }

}