# DEX Upload Release Notes - Version 2.4.1

*Release Date:* 2024-09-20  <br/>
*Version Number:* Version 2.4.1  <br/>
*[Release Changelog](https://github.com/CDCgov/data-exchange-upload/blob/main/CHANGELOG.md)*  <br/>
*[API Swagger Documentation](https://cdcgov.github.io/data-exchange-upload/)* <br/>
*[DEX Upload API Github Releases](https://github.com/CDCgov/data-exchange-upload/releases)* <br/>
*[DEX Upload API Github Tags](https://github.com/CDCgov/data-exchange-upload/tags)*

## Overview
The purpose of Upload API release version 2.4.1 is to expand the `/info` endpoint to include file delivery status for a given upload ID.  End users will be able to use this endpoint to verify whether or not their file got delivered to the intended destination.  This release also includes improved stability to the in-memory eventing system, as well as configuration changes for the NDLP use case.

## Enhancements
- Delivery status for uploads provided by the `/info` endpoint
- In-memory event system stability improvements

## Reporting Team
- DEX Upload Team Distribution List - DEXUploadAPI@cdc.gov