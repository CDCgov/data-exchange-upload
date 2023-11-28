package cdc.gov.upload.client.model;

import java.util.Date;

public class FileStatus {
    
    public String status;
    public double percent_complete;
    public String file_name;
    public int file_size_bytes;
    public int bytes_uploaded;
    public String tus_upload_id;
    public int time_uploading_sec;
    public Metadata metadata;
    public Date timestamp;

    public String getStatus() {
        return status;
    }
    public void setStatus(String status) {
        this.status = status;
    }
    public double getPercent_complete() {
        return percent_complete;
    }
    public void setPercent_complete(double percent_complete) {
        this.percent_complete = percent_complete;
    }
    public String getFile_name() {
        return file_name;
    }
    public void setFile_name(String file_name) {
        this.file_name = file_name;
    }
    public int getFile_size_bytes() {
        return file_size_bytes;
    }
    public void setFile_size_bytes(int file_size_bytes) {
        this.file_size_bytes = file_size_bytes;
    }
    public int getBytes_uploaded() {
        return bytes_uploaded;
    }
    public void setBytes_uploaded(int bytes_uploaded) {
        this.bytes_uploaded = bytes_uploaded;
    }
    public String getTus_upload_id() {
        return tus_upload_id;
    }
    public void setTus_upload_id(String tus_upload_id) {
        this.tus_upload_id = tus_upload_id;
    }
    public int getTime_uploading_sec() {
        return time_uploading_sec;
    }
    public void setTime_uploading_sec(int time_uploading_sec) {
        this.time_uploading_sec = time_uploading_sec;
    }
    public Metadata getMetadata() {
        return metadata;
    }
    public void setMetadata(Metadata metadata) {
        this.metadata = metadata;
    }
    public Date getTimestamp() {
        return timestamp;
    }
    public void setTimestamp(Date timestamp) {
        this.timestamp = timestamp;
    }    
}
