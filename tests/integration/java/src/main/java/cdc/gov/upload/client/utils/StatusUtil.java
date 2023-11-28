package cdc.gov.upload.client.utils;

import java.io.IOException;

import com.fasterxml.jackson.annotation.JsonAutoDetect;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.JsonMappingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.introspect.VisibilityChecker;

import cdc.gov.upload.client.model.FileStatus;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.Response;
import okhttp3.ResponseBody;

public class StatusUtil {
    
    public static FileStatus getFileStatus(String token, String tguid, String baseUrl) throws Exception {
        try {
            String statusUrl = baseUrl + "/status/" + tguid;
            System.out.println("Status URL: " + statusUrl);

            OkHttpClient okHttpClient = new OkHttpClient();
                        
            Request request = new Request.Builder()
                                         .url(statusUrl)                                         
                                         .addHeader("Authorization", "Bearer " + token)
                                         .get()
                                         .build();

            Response response = okHttpClient.newCall(request).execute(); 

            if(response.code() == 400) {
                
                for(int i = 1; i <=3; i++) {   
                    System.out.println("Status Call Retry - " + i);

                    Thread.sleep(3000 * i);
                    response = okHttpClient.newCall(request).execute(); 

                    if(response.code() == 200) {
                        break;
                    }
                }
            }

            if( response.code() == 200) {
                System.out.println("Status Call Successfull!");

                return getStatusResponse(response);
                
            } else  {

                throw new Exception("Status Call Failed with Response code " + response.code());
            }                       
        } catch (Exception e) {

            e.printStackTrace();            
            throw e;
        }
    }

    private static FileStatus getStatusResponse(Response response)
                throws JsonProcessingException, JsonMappingException, IOException {
            ResponseBody responseBody = response.body(); 

            ObjectMapper objectMapper = new ObjectMapper();
            objectMapper.disable(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES);
            objectMapper.setVisibility(VisibilityChecker.Std.defaultInstance().withFieldVisibility(JsonAutoDetect.Visibility.ANY));

            FileStatus statusResponse = objectMapper.readValue(responseBody.string(), FileStatus.class);

            return statusResponse;
        }
}
