package cdc.gov.upload.client.model;

import java.util.ArrayList;

public class LoginResponse {
    
    public String access_token;
    public String token_type;
    public int expires_in;
    public String refresh_token;
    public String scope;
    public ArrayList<String> resource;

    public String getAccess_token() {
        return access_token;
    }
    public void setAccess_token(String access_token) {
        this.access_token = access_token;
    }
    public String getToken_type() {
        return token_type;
    }
    public void setToken_type(String token_type) {
        this.token_type = token_type;
    }
    public int getExpires_in() {
        return expires_in;
    }
    public void setExpires_in(int expires_in) {
        this.expires_in = expires_in;
    }
    public String getRefresh_token() {
        return refresh_token;
    }
    public void setRefresh_token(String refresh_token) {
        this.refresh_token = refresh_token;
    }
    public String getScope() {
        return scope;
    }
    public void setScope(String scope) {
        this.scope = scope;
    }
    public ArrayList<String> getResource() {
        return resource;
    }
    public void setResource(ArrayList<String> resource) {
        this.resource = resource;
    }    
}
