# DEX Upload Release Notes - Version 1.2.0

*Release Date:* 2023-08-15  <br/>
*Version Number:* Version 1.2.0  <br/>
*[Release Changelog](https://github.com/CDCgov/data-exchange-upload/blob/main/CHANGELOG.md)*  <br/>
*[API Swagger Documentation](https://cdcgov.github.io/data-exchange-upload/)*

## Overview
The purpose of Enterprise Data Exchange (DEX) Upload release version 1.2.0 is to enable new application features that will track and store file upload progress status, provide an API to retrieve status updates of uploads, and to provide a user-based upload authorization workflow derived upon Secure Access Management Services (SAMS) authentication scopes.

## New Features
### Supplemental API - Upload Status
*Dex Upload API Physical Architecture v1.2.0*
![DEX Upload API Physical Architecture v1.2.0](diagrams/DEX%20Upload%20API%20Physical%20Architecture%20v1.2.0.png)

The DEX supplemental Application Program Interface (API) feature is designed to allow DEX Upload users to request and obtain status information regarding the progress of file uploads. 
- Metadata regarding upload progress are captured and passed to a storage account queue.
- A function application initially logs the upload and then subsequently synchronizes progress status as the upload continues to completion; persisted within a Cosmos database.
- An API, when called by an upload user, calls a function application that retrieves the current upload status from the Cosmos database and returns that status to the user that called the API.

### SAMS Scopes Implementation
*Password Grant Flow*
![Password Grant Flow](diagrams/Authentication%20Flow%20between%20DEX%20and%20SAMS%20-%20Password%20Grant%20Flow.png)

The SAMS scopes implementation is designed to allow authorization to DEX Upload functionality based on scopes defined within SAMS. 
- For customers utilizing user authentication tokens provided by SAMS to access the DEX Upload API:
    - DEX Upload API will route the user's provided token via Azure API Management (APIM) to SAMS. 
    - Upon validation of the token authentication and user authorization scope, DEX Upload will allow the requested action to continue (file upload and/or status checks). 
- For customers using SAMS system accounts and password grant to access the DEX Upload API:
    - Upon successful authentication, the system account will be provided with an authorization token that contains scopes to perform file upload and status check actions.

**SAMS Scopes Implemented**
- dex:upload - Authorizes file uploads
- dex:status - Authorizes supplemental API status requests

## Reporting Team
- DEX Upload Team Distribution List - dexuploadapi@cdc.gov