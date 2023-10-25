package gov.cdc.ocio.supplementalapi.model;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonProperty;

@JsonIgnoreProperties(ignoreUnknown = true)
public class Destination {
    @JsonProperty("destination_id")
    public String destinationId;
}
