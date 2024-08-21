# DEX Upload Release Notes - Version 2.3.0

*Release Date:* 2024-08-21  <br/>
*Version Number:* Version 2.3.0  <br/>
*[Release Changelog](https://github.com/CDCgov/data-exchange-upload/blob/main/CHANGELOG.md)*  <br/>
*[API Swagger Documentation](https://cdcgov.github.io/data-exchange-upload/)* <br/>
*[DEX Upload API Github Releases](https://github.com/CDCgov/data-exchange-upload/releases)* <br/>
*[DEX Upload API Github Tags](https://github.com/CDCgov/data-exchange-upload/tags)*

## Overview
The purpose of Enterprise Data Exchange (DEX) Upload release version 2.3.0 is to implement the new PS API report schema, as well as consolidate the service's post processing and delivery capability.  Before this release, the Upload API depended on a separate microservice for performing upload post processing and delivery to the EDAV or Routing storage locations.  This microservice was triggered off of an Azure event hub, that received messages from an Azure event grid when a file upload was complete.  This eventing and post processing capability has been consolidated into the upload tus service itself.  It utilizes tus hooks for post processing, and in-memory queues for event dispatching and handling.  This has resulted in the reduction of the Upload API's cloud footprint within Azure to a single compute instance (currently an App Service) and a Redis cache to handle the distributed file locking.

## Enhancements
- Embedding of file post processing and delivery capabilities into the upload tus service
- v2 PS API reports for improved observability
- Decommission of Upload API function app, Event Hub, and Event Grid

## Reporting Team
- DEX Upload Team Distribution List - DEXUploadAPI@cdc.gov