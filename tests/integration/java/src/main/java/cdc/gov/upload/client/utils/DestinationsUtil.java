package cdc.gov.upload.client.utils;

import java.io.File;
import java.io.FileInputStream;
import java.io.InputStream;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.DeserializationFeature;
import com.fasterxml.jackson.databind.ObjectMapper;

import cdc.gov.upload.client.model.Definition;
import cdc.gov.upload.client.model.Destination;
import cdc.gov.upload.client.model.Field;

public class DestinationsUtil {
    
    public static List<Destination> getAllDestinations(String configsFolder) throws Exception {

        Path path  = Paths.get(configsFolder + "/" + "allowed_destination_and_events.json");
        InputStream in = new FileInputStream(path.normalize().toAbsolutePath().toString());

        ObjectMapper objectMapper = new ObjectMapper();
        objectMapper.disable(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES);  

        List<Destination> destinations = objectMapper.readValue(in, new TypeReference<List<Destination>>() {});

        in.close();
        return destinations;
    }

    public static List<Definition> getMetadtaDefinition(String configsFolder, String fileName) throws Exception {

        Path path  = Paths.get(configsFolder + "/" + fileName);
        InputStream in = new FileInputStream(path.normalize().toAbsolutePath().toString());

        ObjectMapper objectMapper = new ObjectMapper();
        objectMapper.disable(DeserializationFeature.FAIL_ON_UNKNOWN_PROPERTIES);  

        List<Definition> definitions = objectMapper.readValue(in, new TypeReference<List<Definition>>() {});

        in.close();
        return definitions;
    }

    public static Map<String, String> generateMetadataFromConfigs(String destination, String event, File file, List<Definition> definitions) throws Exception {
        
        HashMap<String, String> metadata = new HashMap<>();

        metadata.put("meta_destination_id", destination);
        metadata.put("meta_ext_event", event);        

        for(Definition definition: definitions) {
            
            List<Field> fields = definition.getFields();

            for (Field field : fields) {
                
                String fieldName = field.getFieldname();
                String requiredString = field.getRequired();
                Boolean required = Boolean.parseBoolean(requiredString);

                if(required) {
                    if(!metadata.containsKey(fieldName)) {                        
                        List<String> allowedValues = (List<String>)field.getAllowed_values();
                        
                        if (allowedValues != null ) {
                            metadata.put(fieldName, allowedValues.get(0));
                        } else {
                            if(fieldName.equalsIgnoreCase("filename")){
                                metadata.put(fieldName, file.getName());
                            } else if(fieldName.equalsIgnoreCase("meta_ext_filename")){
                                metadata.put(fieldName, file.getName());
                            } else if(fieldName.equalsIgnoreCase("original_file_timestamp")){
                                metadata.put(fieldName, String.valueOf(file.lastModified()));
                            } else if(fieldName.equalsIgnoreCase("meta_ext_file_timestamp")){
                                metadata.put(fieldName, String.valueOf(file.lastModified()));
                            } else {
                                metadata.put(fieldName, "INTEGRATION-TEST");
                            }
                        }
                    }
                }
            }
        }

        return metadata;
    }

    public static Map<String, String> getMetadata(String destination, String event, File file) throws Exception {

        HashMap<String, String> metadataMap = new HashMap<>();

        if(destination.equalsIgnoreCase("dextesting") &&  event.equalsIgnoreCase("testevent1")) {

            metadataMap.put("meta_destination_id", destination);
            metadataMap.put("meta_ext_event", event);        
            metadataMap.put("filename", file.getName());
            metadataMap.put("meta_ext_source", "INTEGRATION-TEST");
    
        } else if(destination.equalsIgnoreCase("ndlp") &&  event.equalsIgnoreCase("routineImmunization")) {
        
            metadataMap.put("meta_destination_id", "ndlp");
            metadataMap.put("meta_ext_source", "IZGW");
            metadataMap.put("meta_ext_sourceversion", "V2022-12-31");
            metadataMap.put("meta_ext_event", "routineImmunization");           
            metadataMap.put("meta_ext_filename", "TEST-FILE-3.zip");
            metadataMap.put("meta_ext_submissionperiod", "2023Q1");
            metadataMap.put("meta_schema_version", "1.0");
            metadataMap.put("izgw_route_id", "dex-stg");
            metadataMap.put("izgw_ipaddress", "127.0.0.1");
            metadataMap.put("izgw_filesize", "1781578");
            metadataMap.put("izgw_path", "/upload/files/f394848234438a40878125e990adfdc7");
            metadataMap.put("izgw_uploaded_timestamp", "Mon, 21 Aug 2023 20:00:21 UTC");
            metadataMap.put("meta_ext_entity", "MAA");
            metadataMap.put("meta_username", "integration.testing.izgateway.org");
            metadataMap.put("meta_ext_objectkey", "5c70b304-9a07-3329-8cac-a64a5dcba380");

        } else {

            throw new Exception("No metadta found for Destination: " + destination + " Event: " + event);
        }

        return metadataMap;
    }
}
