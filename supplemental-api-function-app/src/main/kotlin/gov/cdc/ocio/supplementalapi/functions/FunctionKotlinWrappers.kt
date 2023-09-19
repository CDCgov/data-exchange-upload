package gov.cdc.ocio.supplementalapi.functions
import com.microsoft.azure.functions.*
import com.microsoft.azure.functions.HttpMethod
import com.microsoft.azure.functions.annotation.AuthorizationLevel
import com.microsoft.azure.functions.annotation.BindingName
import com.microsoft.azure.functions.annotation.FunctionName
import com.microsoft.azure.functions.annotation.HttpTrigger
import java.util.Optional

class FunctionKotlinWrappers {
    @FunctionName("HealthCheck")
    fun healthCheck(
        @HttpTrigger(
            name = "req",
            methods = [HttpMethod.GET],
            route = "status/health",
            authLevel = AuthorizationLevel.ANONYMOUS
        ) request: HttpRequestMessage<*>,
        context: ExecutionContext
    ): HttpStatus {
        return HealthCheckFunction().run(request, context)
    }

    @FunctionName("StatusForTguid")
    fun statusForTguid(
        @HttpTrigger(
            name = "req",
            methods = [HttpMethod.GET],
            route = "status/{tguid}",
            authLevel = AuthorizationLevel.FUNCTION
        ) request: HttpRequestMessage<Optional<String>>,
        @BindingName("tguid") tguid: String,
        context: ExecutionContext
    ): HttpResponseMessage {
        return StatusForTguidFunction().run(request, tguid, context)
    }

    @FunctionName("StatusForDestination")
    fun statusForDestination(
        @HttpTrigger(
            name = "req",
            methods = [HttpMethod.GET],
            route = "status/destination/{destinationName}",
            authLevel = AuthorizationLevel.FUNCTION
        ) request: HttpRequestMessage<Optional<String>>,
        @BindingName("destinationName") destinationName: String,
        context: ExecutionContext
    ): HttpResponseMessage {
        return StatusForDestinationFunction().run(request, destinationName, context)
    }
}
