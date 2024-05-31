# DEX Upload Release Notes - Version 2.1.0

*Release Date:* 2024-05-30 <br/>
*Version Number:* Version 2.1.1 <br/>
*Release Changelog*
*[API Swagger Documentation](https://github.com/CDCgov/data-exchange-upload/blob/main/docs/openapi.yml)* <br/>
*[DEX Upload API Github Releases](https://github.com/CDCgov/data-exchange-upload/releases)* <br/>
*[DEX Upload API Github Tags](https://github.com/CDCgov/data-exchange-upload/tags)*

## Overview
Upload API v2 brings several enhancements to existing capabilities, and a few new features.  Note that this release does not include any breaking changes.  Current users do not need to take any action to keep their upload clients running successfully.

## Enhancements
- Upgrade to the v2 TUS resumable upload protocol.  This brings more resilient file uploading, and better error messages.
- Support backwards-compatible upload config filenames.  This allows senders migrating to v2 sender manifests to use different filenames for their v2 upload config JSON file.
- Improved reporting to the Processing Status API.  This includes several new report types, but also a full migration to sending report messages to an Azure Service Bus instead of over HTTP.  This improves observability into files being uploaded and processed.
- The Upload API is no longer sending trace information to the Processing Status API.
- Improved logging.  This improves debugging and troubleshooting.
- Support for CELR, NRSS, and EHDI programs.

## New Features
- Info endpoint.  This is a new HTTP endpoint where authenticated users can send a GET request to /upload/info/{uploadID} and get a response containing metadata about the file that was uploaded.  The uploadID path parameter is the unique ID given back by our service when an upload is complete.
- Health check endpoint.  This is a new HTTP endpoint where users authenticated users can send a GET request to /upload/health and check the overall health of the Upload API service, and the other critical services that it depends on.

## Reporting Team
- DEX Upload Team Distribution List - dexuploadapi@cdc.gov