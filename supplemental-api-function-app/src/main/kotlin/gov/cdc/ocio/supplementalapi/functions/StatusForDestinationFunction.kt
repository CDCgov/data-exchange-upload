package gov.cdc.ocio.supplementalapi.functions

import com.azure.cosmos.models.CosmosQueryRequestOptions
import com.microsoft.azure.functions.ExecutionContext
import com.microsoft.azure.functions.HttpRequestMessage
import com.microsoft.azure.functions.HttpResponseMessage
import com.microsoft.azure.functions.HttpStatus
import gov.cdc.ocio.supplementalapi.cosmos.CosmosClientManager
import gov.cdc.ocio.supplementalapi.exceptions.BadRequestException
import gov.cdc.ocio.supplementalapi.model.Item
import gov.cdc.ocio.supplementalapi.model.UploadStatus
import gov.cdc.ocio.supplementalapi.model.UploadsStatus
import java.text.ParseException
import java.text.SimpleDateFormat
import java.util.*
import java.util.logging.Level


class StatusForDestinationFunction {

    fun run(
        request: HttpRequestMessage<Optional<String>>,
        destinationName: String,
        context: ExecutionContext
    ): HttpResponseMessage {

        val logger = context.logger

        logger.info("HTTP trigger processed a ${request.httpMethod.name} request.")
        logger.info("destination name = $destinationName")

        val dateStart = request.queryParameters["date_start"]
        val dateEnd = request.queryParameters["date_end"]
        val pageSize = request.queryParameters["page_size"]
        val pageNumber = request.queryParameters["page_number"]
        val extEvent = request.queryParameters["ext_event"]
        val sortBy = request.queryParameters["sort_by"]
        val sortOrder = request.queryParameters["sort_order"]

        val pageSizeAsInt = try {
            getPageSize(pageSize)
        } catch (ex: BadRequestException) {
            return request
                .createResponseBuilder(HttpStatus.BAD_REQUEST)
                .body(ex.localizedMessage)
                .build()
        }

        val cosmosClient = CosmosClientManager.getCosmosClient()
        val cosmosDB = cosmosClient.getDatabase("UploadStatus")
        val container = cosmosDB.getContainer("Items")

        val sqlQuery = StringBuilder()
        sqlQuery.append("from Items t where t.meta_destination_id = '$destinationName'")
        extEvent?.run {
            sqlQuery.append(" and t.meta_ext_event = '$extEvent'")
        }

        dateStart?.run {
            try {
                val dateStartEpochSecs = getEpochFromDateString(dateStart, "date_start")
                sqlQuery.append(" and t._ts >= $dateStartEpochSecs")
            } catch (e: BadRequestException) {
                logger.log(Level.SEVERE, e.localizedMessage)
                return request
                    .createResponseBuilder(HttpStatus.BAD_REQUEST)
                    .body(e.localizedMessage)
                    .build()
            }
        }
        dateEnd?.run {
            try {
                val dateEndEpochSecs = getEpochFromDateString(dateEnd, "date_end")
                sqlQuery.append(" and t._ts < $dateEndEpochSecs")
            } catch (e: BadRequestException) {
                logger.log(Level.SEVERE, e.localizedMessage)
                return request
                    .createResponseBuilder(HttpStatus.BAD_REQUEST)
                    .body(e.localizedMessage)
                    .build()
            }
        }

        val countQuery = "select value count(1) $sqlQuery"
        val count = container.queryItems(
            countQuery, CosmosQueryRequestOptions(),
            Long::class.java
        )
        val totalItems = if (count.count() > 0) count.first().toLong() else -1
        val numberOfPages = (totalItems / pageSizeAsInt + if (totalItems % pageSizeAsInt > 0) 1 else 0).toInt()

        val pageNumberAsInt = try {
            getPageNumber(pageNumber, numberOfPages)
        } catch (ex: BadRequestException) {
            return request
                .createResponseBuilder(HttpStatus.BAD_REQUEST)
                .body(ex.localizedMessage)
                .build()
        }

        sortBy?.run {
            val sortField = when (sortBy) {
                "date" -> "_ts"
                else -> {
                    return request
                        .createResponseBuilder(HttpStatus.BAD_REQUEST)
                        .body("sort_by must be one of the following: [date]")
                        .build()
                }
            }
            var sortOrderVal = DEFAULT_SORT_ORDER
            sortOrder?.run {
                sortOrderVal = when (sortOrder) {
                    "ascending" -> "asc"
                    "descending" -> "desc"
                    else -> {
                        return request
                            .createResponseBuilder(HttpStatus.BAD_REQUEST)
                            .body("sort_order must be one of the following: [ascending, descending]")
                            .build()
                    }
                }
            }
            sqlQuery.append(" order by t.$sortField $sortOrderVal")
        }
        val offset = (pageNumberAsInt - 1) * pageSizeAsInt
        val dataSqlQuery = "select * $sqlQuery offset $offset limit $pageSizeAsInt"
        val items = container.queryItems(
            dataSqlQuery, CosmosQueryRequestOptions(),
            Item::class.java
        )

        val uploadsStatus = UploadsStatus()
        items.forEach { item ->
            uploadsStatus.items.add(UploadStatus.createFromItem(item))
        }

        uploadsStatus.summary.page_number = pageNumberAsInt
        uploadsStatus.summary.page_size = pageSizeAsInt
        uploadsStatus.summary.number_of_pages = numberOfPages
        uploadsStatus.summary.total_items = totalItems

        return request
            .createResponseBuilder(HttpStatus.OK)
            .header("Content-Type", "application/json")
            .body(uploadsStatus)
            .build()
    }

    @Throws(BadRequestException::class)
    private fun getPageSize(pageSize: String?) = run {
        var pageSizeAsInt = DEFAULT_PAGE_SIZE
        pageSize?.run {
            var issue = false
            try {
                pageSizeAsInt = pageSize.toInt()
                if (pageSizeAsInt < MIN_PAGE_SIZE || pageSizeAsInt > MAX_PAGE_SIZE)
                    issue = true
            } catch (e: NumberFormatException) {
                issue = true
            }

            if (issue) {
                throw BadRequestException("\"page_size must be between $MIN_PAGE_SIZE and $MAX_PAGE_SIZE\"")
            }
        }
        pageSizeAsInt
    }

    @Throws(BadRequestException::class)
    private fun getPageNumber(pageNumber: String?, numberOfPages: Int) = run {
        var pageNumberAsInt = DEFAULT_PAGE_NUMBER
        pageNumber?.run {
            var issue = false
            try {
                pageNumberAsInt = pageNumber.toInt()
                if (pageNumberAsInt < MIN_PAGE_NUMBER || pageNumberAsInt > numberOfPages)
                    issue = true
            } catch (e: NumberFormatException) {
                issue = true
            }

            if (issue) {
                throw BadRequestException("page_number must be between $MIN_PAGE_NUMBER and $numberOfPages")
            }
        }
        pageNumberAsInt
    }

    @Throws(BadRequestException::class)
    private fun getEpochFromDateString(dateStr: String, fieldName: String): Long {
        try {
            return sdf.parse(dateStr).time / 1000 // convert to secs from millisecs
        } catch (e: ParseException) {
            throw BadRequestException("Failed to parse $fieldName: $dateStr.  Format should be: $DATE_FORMAT.")
        }
    }

    companion object {
        private const val DATE_FORMAT = "yyyyMMdd'T'HHmmss'Z'"
        private val sdf = SimpleDateFormat(DATE_FORMAT)

        private const val MIN_PAGE_SIZE = 1
        private const val MAX_PAGE_SIZE = 10000
        private const val DEFAULT_PAGE_NUMBER = 1

        private const val MIN_PAGE_NUMBER = 1
        private const val DEFAULT_PAGE_SIZE = 100

        private const val DEFAULT_SORT_ORDER = "asc"
    }
}