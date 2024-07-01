# DEX Upload Release Notes - Version 2.2.0

*Release Date:* 2024-06-24  <br/>
*Version Number:* Version 2.2.0  <br/>
*[Release Changelog](https://github.com/CDCgov/data-exchange-upload/blob/main/CHANGELOG.md)*  <br/>
*[API Swagger Documentation](https://cdcgov.github.io/data-exchange-upload/)* <br/>
*[DEX Upload API Github Releases](https://github.com/CDCgov/data-exchange-upload/releases)* <br/>
*[DEX Upload API Github Tags](https://github.com/CDCgov/data-exchange-upload/tags)*

## Overview
The purpose of Enterprise Data Exchange (DEX) Upload release version 2.2.0 is to release an improvement to how the Upload API prevents file corruption when the service horizontally scales.  This improvement is implemented in this release via a distributed file locker that uses a Redis cache to keep track of which Tus instance is currently working on a file, and makes sure another Tus instance doesn't work on the same file.  Before this improvement, we were relying on Azure's internal blob storage implementation to prevent clusters of tus servers from corrupting files.  This implementation is obscured from us and completely out of our control.  This custom distributed lock with Redis gives us full control, and has even been adopted by the Tusd maintainers.

## Enhancements
- Improved reliability of horizontal scaling via a custom file locking mechanism that uses an external Redis cache.

## Reporting Team
- DEX Upload Team Distribution List - dexuploadapi@cdc.gov