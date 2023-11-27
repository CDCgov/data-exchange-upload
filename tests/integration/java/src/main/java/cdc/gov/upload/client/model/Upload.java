package cdc.gov.upload.client.model;

public class Upload {

    private String destination;
    private String event;
    private String fileName;
    private String tguid;
    private String uploadStatus;

    public String getDestination() {
        return destination;
    }
    public void setDestination(String destination) {
        this.destination = destination;
    }
    public String getEvent() {
        return event;
    }
    public void setEvent(String event) {
        this.event = event;
    }
    public String getFileName() {
        return fileName;
    }
    public void setFileName(String fileName) {
        this.fileName = fileName;
    }
    public String getTguid() {
        return tguid;
    }
    public void setTguid(String tguid) {
        this.tguid = tguid;
    }
    public String getUploadStatus() {
        return uploadStatus;
    }
    public void setUploadStatus(String uploadStatus) {
        this.uploadStatus = uploadStatus;
    }

    
}
