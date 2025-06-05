# Upload Configuration

## Overview
In order to establish observability and accommodate program-specific requirements, the upload service utilizes configuration files that define expected metadata values that accompany uploads and determine file delivery logistics.

Each data stream and file type `route` has a configuration file defined. Fields and values defined in these files are utilized by the upload service to validate metadata for accuracy and to deliver files to targets in a manner expected by receiving CDC programs.

## Configuration File Format
Data stream metadata configurations are saved in a JSON format and reside in configured storage for utilization by upload services.

### Configuration Objects
| Field | Type | Description | 
| --- | --- | --- |
| metadata_config | object | Object containing metadata fields and field requirements utilized. This object contains an array of object fields, representing common and custom metadata fields.  |
| copy_config | object | Object containing file delivery information. |

### Object Fields - *metadata_config*
| Field | Type | Description | 
| --- | --- | --- |
| field_name | string | The key name of the field. |
| required | array | Determines requirement status of the field. |
| allowed_values | array | Determines specific values that are accepted; `null` values indicate that any values are accepted.  |
| description | string | Describes the purpose of the field. |

### Object Fields - *copy_config*
| Field | Type | Description | 
| --- | --- | --- |
| filename_suffix | string | Optional field that determines if the filename delivered will have the PHDO upload-id appended to it. |
| folder_structure | string enum | Required field that determines how the file is organized in the destination delivery target. Valid values are `date_YYYY`, `date_YYYY_MM`, `date_YYYY_MM_DD`, `date_YYYY_MM_DD_HH`, and `root`. |
| path_template | string | Optional field that determines to where files are delivered. Values align to paths specific to defined targets. |
| targets | array of strings | Required field that determines to where files are delivered. Values must align to a value in a delivery configuration yml file. |

### Sample Configuration
```json
{
	"metadata_config": {
		"fields": [
			{
				"field_name": "data_stream_id",
				"required": true,
				"allowed_values": [
					"allowed_value_1","allowed_value_2"
				],
				"description": "Data stream identifier; the highest taxonomical designation for a given collection of data to be uploaded."
			},
			{
				"field_name": "data_stream_route",
				"required": true,
				"allowed_values": [
					"allowed_value_1","allowed_value_2"
				],				
				"description": "Second level taxonomical designation for a given collection of data to be uploaded. This value is typically designated to reference a particular file format type."
			},
			{
				"field_name": "sender_id",
				"required": true,
				"allowed_values": [
					"allowed_value_1","allowed_value_2"
				],				
				"description": "Identifier for the submitter of the data."
			},
			{
				"field_name": "data_producer_id",
				"required": true,
				"allowed_values": [
					"allowed_value_1","allowed_value_2"
				],				
				"description": "Identifier for the producer of the submitted data."
			},
			{
				"field_name": "jurisdiction",
				"required": true,
				"allowed_values": [
					"allowed_value_1","allowed_value_2"
				],				
				"description": "Geographical location of the submitted data."
			},
			{
				"field_name": "received_filename",
				"required": true,
				"allowed_values": null,				
				"description": "The name of the file uploaded by the sender."
			},
			{
				"field_name": "custom_metadata_field_example",
				"required": true,
				"allowed_values": [
					"allowed_value_1","allowed_value_2"
				],				
				"description": "Optional custom metadata fields can be added as necessary."
			}
		]
	},
	"copy_config": {
		"filename_suffix": "upload_id",
		"folder_structure": "date_YYYY_MM_DD",
		"path_template": "${EXAMPLE_PATH_TEMPLATE}",
		"targets": [
			"edav"
		]
	}
}
```