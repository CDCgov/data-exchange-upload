# DEX Upload Release Notes - Version 1.6.2

*Release Date:* 2023-04-26  <br/>
*Version Number:* Version 1.6.2  <br/>
*[Release Changelog](https://github.com/CDCgov/data-exchange-upload/blob/main/CHANGELOG.md)*  <br/>
*[API Swagger Documentation](https://cdcgov.github.io/data-exchange-upload/)* <br/>
*[DEX Upload API Github Releases](https://github.com/CDCgov/data-exchange-upload/releases)* <br/>
*[DEX Upload API Github Tags](https://github.com/CDCgov/data-exchange-upload/tags)*

## Overview
The purpose of Enterprise Data Exchange (DEX) Upload release version 1.6.2 is to update upload configurations for data streams, replace filename suffix clock ticks with upload id (tguid), hotfix for Routine Immunization v1 sender manifest config file, and adjust upload filename restrictions.

## New Features
- Update upload filename character restrictions to only disallow forward slashes *(/)*
- Update Routine Immunization v1 sender manifest configuration file to remove incorrect metadata fields
- Add v2 sender manifest configuration files for NRSS and EHDI data streams
- Replace the filename suffix clock tick value with file upload id (tus guid) value

## Reporting Team
- DEX Upload Team Distribution List - dexuploadapi@cdc.gov