# PHDO Upload Release Notes - Version 2.7.1

*Release Date:* 2025-02-25  <br/>
*Version Number:* Version 2.7.1 <br/>
*[Release Changelog](https://github.com/CDCgov/data-exchange-upload/blob/main/CHANGELOG.md)*  <br/>
*[API Swagger Documentation](https://cdcgov.github.io/data-exchange-upload/)* <br/>
*[DEX Upload API Github Releases](https://github.com/CDCgov/data-exchange-upload/releases)* <br/>
*[DEX Upload API Github Tags](https://github.com/CDCgov/data-exchange-upload/tags)*

## Overview
The purpose of Upload API minor release version 2.7.1 is to enable JWT token authentication in the upload server, as well as user session management in the internal UI console.  This release also includes the removal of the eICR data stream.

## Enhancements
- All upload requests are authenticated now with JWT bearer tokens, including PATCH requests.
- OIDC configuration and public key caching for SAMS identity provider.
- User session cookie for Upload UI Console
- MVP login UX with JWT token submittion

## Reporting Team
- PHDO Upload Team Distribution List - DEXUploadAPI@cdc.gov