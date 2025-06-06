# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.8.0] 2025-05-05
- Prometheus metrics for median delivery speed and last successful delivery timestamp
- Export of OTEL trace data to Tempo
- Remove NCIRD datastream information

## [2.7.4] 2025-04-22
- Prometheus metrics for file delivery
- Upload ID logged in the JSON app logs
- Bump the gorilla/csrf package, which requires trusted origins to be explicitly provided
- Bump the redis package

## [2.7.2] 2025-3-31
- Prometheus histogram for upload speed in bytes per second.

## [2.7.1] 2025-3-6
- All upload requests are authenticated now with JWT bearer tokens, including PATCH requests.
- OIDC configuration and public key caching for SAMS identity provider.
- User session cookie for Upload UI Console
- MVP login UX with JWT token submittion

## [2.6.4] 2025-2-25
- Hotfix to prevent overwriting of Azure upload stream errors

## [2.6.2] 2025-2-14
- EHDI configuration removal
- Auto creation of deadletter queue for SQS
- Url decoding for Azure client fix

## [2.6.1] 2024-11-19
- Delivery configuration addition
- Info endpoint delivery name fix
- Info endpoint timestamp precision fix
- Delivery date folder partitioning fix
- Delivery target clean up
- Upload configuration clean up

## [2.5.0] 2024-09-26
- Direct file delivery to CDC Program storage
- Agnostic subfolder delivery path template implementation
- NDLP configuration addition

## [2.4.1] 2024-09-20
- Delivery status addition to the /info endpoint
- In-memory event system stability improvements

## [2.3.0] 2024-08-21
- Embedding of file post processing and delivery capabilities into the upload tus service
- v2 PS API reports for improved observability
- Decommission of Upload API function app, Event Hub, and Event Grid

## [2.2.3] 2024-07-30
- Sender manifest configuration updates
- Global reporting timestamp update to Universal Time Zone (UTC) settings
- Metadata verification error response update to return DEX Upload ID (tguid)
- Internal API Version endpoint addition

## [2.2.0] 2024-06-24
- Improved reliability of horizontal scaling via a custom file locking mechanism that uses an external Redis cache.

## [2.1.2] 2024-06-18
- Version 2 sender manifest configuration for NDLP/Immunization data streams

## [2.1.1] 2024-05-30
- Upgrade to the v2 TUS resumable upload protocol.  This brings more resilient file uploading, and better error messages.
- Support backwards-compatible upload config filenames.  This allows senders migrating to v2 sender manifests to use different filenames for their v2 upload config JSON file.
- Improved reporting to the Processing Status API.  This includes several new report types, but also a full migration to sending report messages to an Azure Service Bus instead of over HTTP.  This improves observability into files being uploaded and processed.
- The Upload API is no longer sending trace information to the Processing Status API.
- Improved logging.  This improves debugging and troubleshooting.
- Support for CELR, NRSS, and EHDI programs.
- Info endpoint.  This is a new HTTP endpoint where authenticated users can send a GET request to /upload/info/{uploadID} and get a response containing metadata about the file that was uploaded.  The uploadID path parameter is the unique ID given back by our service when an upload is complete.
- Health check endpoint.  This is a new HTTP endpoint where authenticated users can send a GET request to /upload/health and check the overall health of the Upload API service, and the other critical services that it depends on.

## [1.6.3] 2024-05-02
- Remove appended tguid values from filenames routed to EDAV storage accounts.

## [1.6.2] 2024-04-23
- Update upload filename character restrictions to only disallow forward slashes (/)
- Update Routine Immunization v1 sender manifest configuration file to remove incorrect metadata fields
- Add v2 sender manifest configuration files for NRSS and EHDI data streams
- Replace the filename suffix clock tick value with file upload id (tus guid) value

## [1.5.0] - 2024-04-23
- Version 2 sender manifest configuration folder structure
- Version 1 sender manifest metadata values to version 2 metadata fields
- File copy functionality to routing storage accounts
- Processing Status API integration
- Retry functionality
- Replay functionality
- Upload configurations for the DAART data stream
- Remove file copy EDAV target from the CELR data stream
- Influenza v1 sender manifest configuration file to correct target routing destination
- Correcting the lower casing of container subfolders created

## [1.4.1] - 2024-03-22
- Renaming of Upload Processor function app
- Infrastructure to support retry/replay functionality
- Application settings to support connection to Processing Status API
- Deployment of blue/green slots to the Upload Processor function app

## [1.2.3] - 2023-01-31
- Added descriptions for the functions in the repo
- Routing Integration: copy to routing changes
- Configured test event to send files to routing
- Metadata update for NDLP accepted values
- Added integration tests suite to upload repo
- Metadata configuration changes for routing
- Integrate app insight Bulk Upload Processor
- Updated log level from error to information
- Unit test for bulk upload
- Metadata definitions added to summary table  ( tus/file-hooks/metadata-verify/definitions/readme.md )

## [1.2.2] - 2023-11-17
- NDLP APL historical data configuration
- Changing all upload configs of NDLP to not append clock ticks to the filename
- Implementation of printing exception stack trace for increased visibility into error details
- Updates to OpenAPI specifications
- Workflow refactoring for upload configurations (CI/CD)
- Addition of workflow for upload configurations (CI/CD)

## [1.2.1] - 2023-09-12
- Fix for ndlp sending data without filename in metadata ( TUS post-receive hook )
- Create upoad root folder in edav ( bulk upload processor )
- Add new NDLP metadata and upload configs ( TUS file hooks: metadata-verify, upload-configs )
- aims-celr configurations ( TUS file hooks: metadata-verify, upload-configs )

## [1.2.0] - 2023-08-15
- Added upload database sync function to log each upload in the persistent storage ( Cosmos DB )
- Added Tus Hooks support to update upload status
- Added Supplemental API Function App to retrieve the status of an upload
- Added SAMS scopes 
- Changes made to the APIM policy, enabling users with specified scopes to access particular functionality
- Added Metadata Configuration changes to support more use cases

## [1.0.2] - 2023-04-27
-  Temporary hotfix to allow NDLP file uploads to work.  IZGW is sending meta_ext_filename, but not filename in the metadata, which is failing in the bulk file upload function app.  Temporary solution is to allow either field, but long-term fix will be to require 'filename' metadata field at time of upload.

## [1.0.1] - 2023-04-14
- HOTFIX to update the NDLP routine immunization required metadata fields to include meta_ext_submissionperiod

## [1.0.0] - 2023-04-04
- Created pre-create tus hook and associated JSON configuration files to support program and event specific metadata validation checks before an upload can proceed
- Created an Azure Function app for supporting bulk file uploads, whose purpose is to combine the tus payload and info files, copy the resultant combined blob file to DEX and EDAV storage accounts
- Github worker files for CI/CD of the tusd pre-create and function app
  
## [0.0.1] - 2022-10-20
- Created initial version of the bulk upload project in Spring Boot/Kotlin/Gradle.
- Added endpoint for health check.
- Added endpoint for multipart file uploads.
