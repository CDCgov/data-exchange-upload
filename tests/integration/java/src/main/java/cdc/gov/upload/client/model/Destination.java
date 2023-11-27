package cdc.gov.upload.client.model;

import java.util.ArrayList;

public class Destination {

    public String destination_id;
    public ArrayList<ExtEvent> ext_events;
    
    public String getDestination_id() {
        return destination_id;
    }
    public void setDestination_id(String destination_id) {
        this.destination_id = destination_id;
    }
    public ArrayList<ExtEvent> getExt_events() {
        return ext_events;
    }
    public void setExt_events(ArrayList<ExtEvent> ext_events) {
        this.ext_events = ext_events;
    }
}
