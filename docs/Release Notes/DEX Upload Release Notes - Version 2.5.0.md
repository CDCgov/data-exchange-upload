# DEX Upload Release Notes - Version 2.5.0

*Release Date:* 2024-09-26  <br/>
*Version Number:* Version 2.5.0 <br/>
*[Release Changelog](https://github.com/CDCgov/data-exchange-upload/blob/main/CHANGELOG.md)*  <br/>
*[API Swagger Documentation](https://cdcgov.github.io/data-exchange-upload/)* <br/>
*[DEX Upload API Github Releases](https://github.com/CDCgov/data-exchange-upload/releases)* <br/>
*[DEX Upload API Github Tags](https://github.com/CDCgov/data-exchange-upload/tags)*

## Overview
The purpose of Upload API release version 2.5.0 is to facilitate direct delivery of file uploads to program storage destinations.  Instead of files being delivered indirectly though a separate routing microservice, the Upload API service will now deliver files directly to CDC program storage.  To enable this, we've decoupled source and destination information from the delivery construct.  This allows you to register sources and destinations separately, which increases flexibilty in configuring where the file is coming from and going to.  In addition, we've increased the flexibilty of where files get dropped in destination storage.  Before, files could be configured to be dropped in subfolders by datastream, route, and date.  Now, the subfolder path can take on any pattern.  This is accomplished through a new `path_template` field that can be specified in a particular upload config.

## Enhancements
- Direct file delivery to CDC program storage
- Agnostic subfolder delivery path template

## Reporting Team
- DEX Upload Team Distribution List - DEXUploadAPI@cdc.gov