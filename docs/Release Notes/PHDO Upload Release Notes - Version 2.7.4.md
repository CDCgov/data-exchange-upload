# PHDO Upload Release Notes - Version 2.7.4

*Release Date:* 2025-04-22  <br/>
*Version Number:* Version 2.7.4 <br/>
*[Release Changelog](https://github.com/CDCgov/data-exchange-upload/blob/main/CHANGELOG.md)*  <br/>
*[API Swagger Documentation](https://cdcgov.github.io/data-exchange-upload/)* <br/>
*[DEX Upload API Github Releases](https://github.com/CDCgov/data-exchange-upload/releases)* <br/>
*[DEX Upload API Github Tags](https://github.com/CDCgov/data-exchange-upload/tags)*

## Overview
The purpose of Upload API minor release version 2.7.4 is to enhance observability metrics, upload server app logs, and patch dependencies for security purposes.

## Enhancements
- Prometheus metrics for file delivery
- Upload ID logged in the JSON app logs

## Updates
- Bump the gorilla/csrf package, which requires trusted origins to be explicitly provided
- Bump the redis package

## Reporting Team
- PHDO Upload Team Distribution List - DEXUploadAPI@cdc.gov