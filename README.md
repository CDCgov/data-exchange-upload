# CDC Data Exchange (DEX) Upload
 
The CDC Data Exchange (DEX) Upload API is an open-source service created to support public health data providers in their effort to share critical public health information with internal CDC Programs. The open-source model allows users to tailor the tool to fit specific data needs. <br/>

The Upload API service is a highly scalable, highly reliable means of transiting files of nearly any type and size from public health partners to the CDC, even when sent over unreliable network connections.

## Upload Services

**Resumable Uploads**: Implementation of the Tus service as a library in a Golang application, facilitating high volume file uploads over HTTP. 

**Authentication**: OAuth provider approval enforcement to ensure security.

**Metadata Verification**: Enforcement of upload submission metadata standards.

**File Delivery**: Routing of uploaded files to configured target destinations.

**File Observability**: Upload lifecycle event tracking, open telemetry traces and metrics, and endpoints for health check, information, and application version. 

**Retry Delivery**: Tool that delivers files to target destinations that uploaded successfully, but were unsuccessful in delivery.

**User Interface**: Facilitates uploads and observability within a user interface.

## Repository Contents

### .github

Utilized for test and CI/CD automation (`/workflows`).

### docs

Release Notes (`/Release Notes`), OAS3 API Documentation, and other relevant documentation.

### reupload-tool-go

Retry functionality that delivers files to target destinations that have uploaded successfully, but were unsuccessful in delivery.
- [Delivery Retry README](https://github.com/CDCgov/data-exchange-upload/blob/main/reupload-tool-go/README.md)

### tests

Smoke test suites leveraging kotlin (`/smoke/kotlin`) and playwright (`/smoke/playwright`).
- [Kotlin Smoke Test README](https://github.com/CDCgov/data-exchange-upload/blob/main/tests/smoke/kotlin/README.md)
- [Playwright Smoke Test README](https://github.com/CDCgov/data-exchange-upload/blob/main/tests/smoke/playwright/README.md)

Load testing tool enabling high volumes of file uploads (`/bad-uploader`).
- [Load Testing Tool README](https://github.com/CDCgov/data-exchange-upload/blob/main/tests/bad-uploader/readme.md)

### upload-configs

Configuration files containing manifest schema values and routing delivery details for specific data stream use cases. JSON configuration files are utilized to verify metadata accommpanying uploads and to determine file delivery to specified target locations.
- [Upload Configs README](https://github.com/CDCgov/data-exchange-upload/blob/main/upload-configs/README.md)

### upload-reports

Go application that fetches data from the [Processing Status GraphQL API](https://github.com/CDCgov/data-exchange-processing-status), generates a CSV report, and optionally uploads it to S3.
- [Upload Reports README](https://github.com/CDCgov/data-exchange-upload/blob/main/upload-reports/README.md)

### upload-scripts

Go script that connects to Azure storage, lists blobs within a specific container, and deletes them.
- [Upload Scripts README](https://github.com/CDCgov/data-exchange-upload/blob/main/upload-scripts/README.md)

### upload-server

Upload server functionality leveraging Tus v2 capabilities written in Golang. Capabilities include resumable file uploads, metadata verification, event routing, observability endpoints, file delivery, distributed file locking, OAuth token verification, user interface, unit testing, and integration testing.
- [Upload Server README](https://github.com/CDCgov/data-exchange-upload/blob/main/upload-server/readme.md)

## TUS Protocol

This repository is using the TUS resumable upload protocol: [https://tus.io/](https://tus.io/), and reference implementation: [https://github.com/tus/tusd](https://github.com/tus/tusd)

## Example Usage

Example clients, for back-end or browser (front-end), to upload files: [https://github.com/CDCgov/data-exchange-api-examples](https://github.com/CDCgov/data-exchange-api-examples)

## Future Improvements

- Upload routing configuration defined within the upload server; removes JSON configuration file dependency

## Public Domain Standard Notice

This repository constitutes a work of the United States Government and is not subject to domestic copyright protection under 17 USC § 105. This repository is in the public domain within the United States, and copyright and related rights in the work worldwide are waived through the [CC0 1.0 Universal public domain dedication](https://creativecommons.org/publicdomain/zero/1.0/). All contributions to this repository will be released under the CC0 dedication. By submitting a pull request you are agreeing to comply with this waiver of copyright interest.

## License Standard Notice

The repository utilizes code licensed under the terms of the Apache Software License and therefore is licensed under ASL v2 or later. <br/>

This source code in this repository is free: you can redistribute it and/or modify it under the terms of the Apache Software License version 2, or (at your option) any later version. <br/>

This source code in this repository is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the Apache Software License for more details. <br/>

You should have received a copy of the Apache Software License along with this program. If not, see [http://www.apache.org/licenses/LICENSE-2.0.html](http://www.apache.org/licenses/LICENSE-2.0.html). <br/>

The source code forked from other open source projects will inherit its license.

## Privacy Standard Notice

This repository contains only non-sensitive, publicly available data and information. All material and community participation is covered by the [Disclaimer](https://github.com/CDCgov/template/blob/master/DISCLAIMER.md) and [Code of Conduct](https://github.com/CDCgov/template/blob/mastercode-of-conduct.md). For more information about CDC's privacy policy, please visit [http://www.cdc.gov/other/privacy.html](https://www.cdc.gov/other/privacy.html).

## Contributing Standard Notice

Anyone is encouraged to contribute to the repository by [forking](https://help.github.com/articles/fork-a-repo) and submitting a pull request. (If you are new to GitHub, you might start with a [basic tutorial](https://help.github.com/articles/set-up-git).) By contributing to this project, you grant a world-wide, royalty-free, perpetual, irrevocable, non-exclusive, transferable license to all users under the terms of the [Apache Software License v2](http://www.apache.org/licenses/LICENSE-2.0.html) or later. <br/>

All comments, messages, pull requests, and other submissions received through CDC including this GitHub page may be subject to applicable federal law, including but not limited to the Federal Records Act, and may be archived. Learn more at [http://www.cdc.gov/other/privacy.html](http://www.cdc.gov/other/privacy.html).

## Records Management Standard Notice

This repository is not a source of government records, but is a copy to increase collaboration and collaborative potential. All government records will be published through the [CDC web site](http://www.cdc.gov).

## Additional Standard Notices

Please refer to [CDC's Template Repository](https://github.com/CDCgov/template)for more information about [contributing to this repository](https://github.com/CDCgov/template/blob/master/CONTRIBUTING.md), [public domain notices and disclaimers](https://github.com/CDCgov/template/blob/master/DISCLAIMER.md), and [code of conduct](https://github.com/CDCgov/template/blob/master/code-of-conduct.md).