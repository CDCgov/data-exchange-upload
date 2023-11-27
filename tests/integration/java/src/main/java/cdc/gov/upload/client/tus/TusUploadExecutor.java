package cdc.gov.upload.client.tus;

import java.io.File;
import java.io.IOException;
import java.net.URL;
import java.util.HashMap;
import java.util.Map;
import io.tus.java.client.ProtocolException;
import io.tus.java.client.TusClient;
import io.tus.java.client.TusExecutor;
import io.tus.java.client.TusURLMemoryStore;
import io.tus.java.client.TusUpload;
import io.tus.java.client.TusUploader;

public class TusUploadExecutor extends  TusExecutor {

    private TusClient client = null;
    private TusUpload upload = null;
    private String tguid = null;
    
    public void initiateUpload(String token, String baseUrl, File file, Map<String, String> metadata) throws Exception {

        try {
            client = new TusClient();
            upload = new TusUpload(file);

            // Configure tus HTTP endpoint. This URL will be used for creating new uploads
            client.setUploadCreationURL(new URL(baseUrl + "/upload"));            

            // Enable resumable uploads by storing the upload URL in memory
            client.enableResuming(new TusURLMemoryStore());
            
            setHeaders(token);

            upload.setMetadata(metadata);
            
            makeAttempts();

        } catch (Exception e) {

            e.printStackTrace();
            throw e;
        }
    }

    public String getTguid() {
        return tguid;
    }

    @Override
    protected void makeAttempt() throws ProtocolException, IOException {

        // First try to resume an upload. If that's not possible we will create a new
        // upload and get a TusUploader in return. This class is responsible for opening
        // a connection to the remote server and doing the uploading.
        TusUploader uploader = client.resumeOrCreateUpload(upload);

        // Upload the file in chunks of 1KB sizes.
        uploader.setChunkSize(1024 * 1024);

        // Upload the file as long as data is available. Once the
        // file has been fully uploaded the method will return -1
        do {
            // Calculate the progress using the total size of the uploading file and
            // the current offset.
            long totalBytes = upload.getSize();
            long bytesUploaded = uploader.getOffset();
            double progress = (double) bytesUploaded / totalBytes * 100;
            
            System.out.printf("Upload at %06.2f%%.\n", progress);                        
        } while (uploader.uploadChunk() > -1);

        // Allow the HTTP connection to be closed and cleaned up
        uploader.finish();
        
        parseTguid(uploader.getUploadURL().toString());
    }

    private void parseTguid(String uploadUrl) {

        String[] url = uploadUrl.split("/");
        tguid = url[url.length-1].trim();
    }    

    private void setHeaders(String token) {

        HashMap<String, String> headerMap = new HashMap<>();

        // Both of these are necessary to work around the 411 issue.
        System.setProperty("sun.net.http.allowRestrictedHeaders", "true");
        headerMap.put("Content-Length","0");
        
        headerMap.put("Authorization", "Bearer " + token);

        client.setHeaders(headerMap);
    }
}
