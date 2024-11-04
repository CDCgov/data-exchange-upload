# DEX Upload Release Notes - Version 1.2.1

*Release Date:* 2023-09-12  <br/>
*Version Number:* Version 1.2.1  <br/>
*[Release Changelog](https://github.com/CDCgov/data-exchange-upload/blob/main/CHANGELOG.md)*  <br/>
*[API Swagger Documentation](https://cdcgov.github.io/data-exchange-upload/)*

## Overview
The purpose of Enterprise Data Exchange (DEX) Upload release version 1.2.1 is to enable new features that will enhance file uploads for National Data Lake Platform (NDLP), Routine Immunization (RI) and Influenza reporting and to enable uploads for Association of Public Health Laboratories (APHL) Informatics Messaging Services (AIMS), COVID Electronic Lab Reporting (CELR).

## New Features
- Tus Post-Receive Hook 
    - Fix for NDLP sending data without filename in metadata
- Bulk Upload Processor Function App 
    - Create upload root folder in EDAV storage
    - NDLP filename changes to overwrite incoming file if uploaded again 
- Tus File Hooks, Metadata-Verify & Upload-Configs 
    - Add new NDLP metadata and upload configurations
    - Add AIMS CELR metadata and upload configurations

## Reporting Team
- DEX Upload Team Distribution List - dexuploadapi@cdc.gov