package gov.cdc.ocio.supplementalapi;

import java.util.*;
import com.microsoft.azure.functions.annotation.*;
import com.microsoft.azure.functions.*;
import gov.cdc.ocio.supplementalapi.functions.StatusForDestinationFunction;
import gov.cdc.ocio.supplementalapi.functions.StatusForTguidFunction;

public class FunctionJavaWrappers {

    @FunctionName("StatusForTguid")
    public HttpResponseMessage statusForTguid(
            @HttpTrigger(
                    name = "req",
                    methods = {HttpMethod.GET},
                    route = "status/{tguid}",
                    authLevel = AuthorizationLevel.FUNCTION) HttpRequestMessage<Optional<String>> request,
            @BindingName("tguid") String tguid,
            final ExecutionContext context) {
        return new StatusForTguidFunction().run(request, tguid, context);
    }

    @FunctionName("StatusForDestination")
    public HttpResponseMessage statusForDestination(
            @HttpTrigger(
                    name = "req",
                    methods = {HttpMethod.GET},
                    route = "status/destination/{destinationName}",
                    authLevel = AuthorizationLevel.FUNCTION) HttpRequestMessage<Optional<String>> request,
            @BindingName("destinationName") String destinationName,
            final ExecutionContext context) {
        return new StatusForDestinationFunction().run(request, destinationName, context);
    }

}
