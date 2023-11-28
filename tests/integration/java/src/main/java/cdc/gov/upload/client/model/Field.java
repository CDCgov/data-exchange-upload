package cdc.gov.upload.client.model;

import java.util.ArrayList;

public class Field {

    public String fieldname;
    public ArrayList<String> allowed_values;
    public String required;
    public String description;

    public String getFieldname() {
        return fieldname;
    }
    public void setFieldname(String fieldname) {
        this.fieldname = fieldname;
    }
    public ArrayList<String> getAllowed_values() {
        return allowed_values;
    }
    public void setAllowed_values(ArrayList<String> allowed_values) {
        this.allowed_values = allowed_values;
    }
    public String getRequired() {
        return required;
    }
    public void setRequired(String required) {
        this.required = required;
    }
    public String getDescription() {
        return description;
    }
    public void setDescription(String description) {
        this.description = description;
    }   
}
