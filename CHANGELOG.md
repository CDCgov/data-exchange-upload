# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


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
