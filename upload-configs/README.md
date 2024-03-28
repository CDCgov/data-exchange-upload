# Upload Configuration

## Overview
As files are copied into the DEX and EDAV storage accounts, various programs may have different requirements on what metadata field is used for the `filename`, whether the `filename` is guaranteed unique by adding a dynamic suffix, the subfolder structure the files are placed in the container, etc.

To make these settings configurable we have "upload configurations", available for each combination of `destination_id` and `ext_event`.

## Details
A JSON file format is used to setup the upload configurations.  The JSON files reside in a container on the DEX storage and is used by the bulk file upload function app.

### JSON File Format

`metadata_config` and `copy_config` are two main objects in the JSON file. 

| Field | Type | Description | 
| --- | --- | --- |
| metadata_config | object | Object containing information about common and use-case specific metadata fields for a given use-case |
| copy_config | object | Object containing information about how and where a file is to be copied to for a given use-case |



metadata_config
| Field | Type | Description | 
| --- | --- | --- |
| version | string | Version number representing the metadata version.  Ex: 1.0, 2.0 |
| fields | array | Array of objects for each common and use-case specific metadata field |



metadata <em>fields</em>
| Field | Type | Description | 
| --- | --- | --- |
| field_name | string | The key name of the field |
| required | array | Whether or not the field is required for the given use-case |
| description | string | Description of the purpose of the field |
| compat_field_name | string | Name of the previous-version compatible field name |


copy_config
| Field | Type | Description | 
| --- | --- | --- |
| filename_suffix | string enum | Determines the suffix that gets appended to the filename when it is copied to destination storage accounts.  Valid values are “clock_ticks” or null |
| folder_structure | string enum | Determines how the file is organized in the destination storage account.  Valid values are “date_YYYY_MM_DD”  and “root” |
| targets | array of strings | Determines where the file copies to, either EDAV or routing |


Example 1:
```json
{
	"metadata_config": {
		"version": "1.0",
		"fields": [
			{
				"field_name": "meta_ext_objectkey",
				"required": true,
				"description": "This field is used to track back to the source objectid.",
				"compat_field_name": null
			},
			{
				"field_name": "meta_ext_filename",
				"required": true,
				"description": "The name of the file submitted.",
				"compat_field_name": null
			},
			{
				"field_name": "meta_ext_file_timestamp",
				"required": true,
				"description": "The timestamp on the source for file last modified.",
				"compat_field_name": null
			},
			{
				"field_name": "meta_username",
				"required": true,
				"description": "Username of user or system name who submitted the file.",
				"compat_field_name": null
			},
			{
				"field_name": "meta_ext_filestatus",
				"required": true,
				"description": "The file status in the source system.",
				"compat_field_name": null
			},
			{
				"field_name": "meta_ext_uploadid",
				"required": true,
				"description": "The uploadid in the system source.",
				"compat_field_name": null
			},
			{
				"field_name": "meta_program",
				"required": "false",
				"description": "The program source.",
				"compat_field_name": null
			},
			{
				"field_name": "meta_ext_source",
				"required": true,
				"description": "The source system.",
				"compat_field_name": null
			},
			{
				"field_name": "meta_organization",
				"required": true,
				"description": "The organization from the source system.",
				"compat_field_name": null
			}
		]
	},
	"copy_config": {
		"filename_suffix": "clock_ticks",
		"folder_structure": "date_YYYY_MM_DD",
		"targets": [
			"routing"
		]
	}
}
```
In this example, note the `compat_field_name` is null for all metadata fields in V1. 

## Setup Steps

1. Create a file for the `data_stream_id` and `data_stream_route` with the format: ```{data_stream_id}-{data_stream_route}.json```. For example, "```ndlp-routineImmunization.json```".
2. Connect to the DEX storage account for the environment.
3. Upload the file to the DEX container, ```upload-configs```.