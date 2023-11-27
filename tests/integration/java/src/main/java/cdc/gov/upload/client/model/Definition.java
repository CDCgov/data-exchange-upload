package cdc.gov.upload.client.model;

import java.util.ArrayList;

public class Definition {

    public String schema_version;
    public ArrayList<Field> fields;

    public String getSchema_version() {
        return schema_version;
    }
    public void setSchema_version(String schema_version) {
        this.schema_version = schema_version;
    }
    public ArrayList<Field> getFields() {
        return fields;
    }
    public void setFields(ArrayList<Field> fields) {
        this.fields = fields;
    }
}
