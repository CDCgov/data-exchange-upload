package cdc.gov.upload.client.utils;

import java.nio.charset.Charset;
import java.nio.charset.StandardCharsets;

import com.fasterxml.jackson.annotation.JsonAutoDetect;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.introspect.VisibilityChecker;

import cdc.gov.upload.client.model.LoginResponse;
import okhttp3.FormBody;
import okhttp3.OkHttpClient;
import okhttp3.Request;
import okhttp3.RequestBody;
import okhttp3.Response;
import okhttp3.ResponseBody;

public class LoginUtil {

    public static String getToken(String username, String password, String baseUrl) throws Exception {
        try {
            String loginUrl = baseUrl + "/oauth";
            System.out.println("Login URL: " + loginUrl);

            OkHttpClient okHttpClient = new OkHttpClient();
                        
            Charset charset = Charset.forName(StandardCharsets.UTF_8.name());
            RequestBody requestBody = new FormBody.Builder(charset)
                                                  .add("username", username)
                                                  .add("password", password)
                                                  .build();

            Request request = new Request.Builder()
                                         .url(loginUrl)
                                         .post(requestBody)
                                         .build();
           

            Response response = okHttpClient.newCall(request).execute();

            if( response.code() == 200) {
                System.out.println("Login Successfull!");

                ResponseBody responseBody = response.body(); 

                ObjectMapper objectMapper = new ObjectMapper();
                objectMapper.disable(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES);
                objectMapper.setVisibility(VisibilityChecker.Std.defaultInstance().withFieldVisibility(JsonAutoDetect.Visibility.ANY));                

                LoginResponse loginResponse = objectMapper.readValue(responseBody.string(), LoginResponse.class);

                return loginResponse.getAccess_token();

            } else {

                throw new Exception("Login Call Failed with Rsponse code " + response.code());
            }             
        } catch (Exception e) {

            e.printStackTrace();            
            throw e;
        }
    }
}