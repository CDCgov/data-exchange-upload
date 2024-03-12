# Upload Configuration

## Overview
As files are copied into the DEX and EDAV storage accounts different programs may have different requirements on what metadata field is used for the filename, whether the filename is guaranteed unique by adding a dynamic suffix, the subfolder structure the files are placed in the container, etc.

To make these settings configurable we have "upload configurations", available for each combination of destination_id and ext_event.

## Details
A JSON file format is used to setup the upload configurations.  The JSON files reside in a container on the DEX storage and is used by the bulk file upload function app.

### JSON File Format
```
{
   "FilenameMetadataField": "e.g. meta_ext_filename",
   "FilenameSuffix": "none | clock_ticks",
   "FolderStructure": "root | path | date_YYYY_MM_DD",
   "FixedFolderPath": "{subfolder1/subfolder2}"
}
```

- ```FilenameMetadataField``` - name of metadata field to use for the destination filename.
- ```FilenameSuffix``` - Options are ```none```, ```clock_ticks```, ```date_YYYY_MM_DD```
  - ```none``` - No file suffix added
  - ```clock_ticks``` - Adds the current time at time of upload as ticks to the filename.
    Example: ```{filename}_638218229220000000.ext```
- ```FolderStructure``` - 
  - ```root``` - No file suffix added
  - ```path``` - Puts the file in a fixed path
  - ```date_YYYY_MM_DD``` - Puts file in subfolders according to the data of upload in the format YYY/MM/DD where YYYY, MM and DD are subfolders.
- ```FixedFolderPath``` - Only necessary to populate if ```FolderStructure``` is set to ```path```, which can be a single folder or series of subfolders.

Example 1:
```
{
   "FilenameMetadataField": "original_filename",
   "FilenameSuffix": "none",
   "FolderStructure": "path",
   "FixedFolderPath": "dex-routing"
}
```
In this example, we will use ```original_filename``` of the metadata to name the files in DEX and EDAV storage containers.  It will have no suffix appended to the filename.  The files will be placed in a fixed subfolder of the container named, ```dex-routing```.

## Setup Steps

1. Create a file for the desination_id and ext_event with the format: ```{destination_id}-{ext_event}.json```.  For example, "```ndlp-routineImmunization.json```".
2. Connext to the DEX storage account for the environment.
3. Upload the file to the dex container, ```upload-configs```.