# CDC Data Exchange (DEX) Upload

This repository manages the CDC Data Exchange (DEX) Upload functionalities.  It is the entrypoint of files into DEX.

## Services

1. **upload-service**: An implementation of the Tus service as a library in a Golang application.  This service is responsible for facilitating file uploads over HTTP.  It also includes a suite of upload lifecycle hooks for establishing observability into the upload process.

## Functions

1. **Bulk File Upload**: blob-created events from the DEX storage account (dataexchange + env) by subscribing to the "tus-upload-complete" event hub.
  It copies files from the raw upload container to the DEX container, organizing them into folders based on accompanying metadata. The metadata configuration is
  used to determine the appropriate final destination, either EDAV or DEX routing containers.


### Repository Structure

#### Github Actions

`/.github/workflows` - tests and CI/CD

#### file-hooks

`tus/file-hooks` hooks employed include:

- pre-create
- post-create
- post-finish
- post-receive

#### upload-configs

`upload-configs` holds configuration for metadata and file routing for each supported use case in JSON format.  These files follow a `{use case}-{use case category}.json` naming convention.  Where the use case represents the data stream, and the use case category represents the data stream route.  For example, AIMS-CELR is a CDC program and supported use case within DEX.  They have one use case ID, `aims-celr`, which represents their data stream ID.  Within this use case, they have two categories: `csv` and `hl7`, which represent the two data stream routes for this program.  To configure AIMS-CELR for the Upload API, we create two JSON files for them, namely `aims-celr-csv.json` and `aims-celr-hl7.json`.  This pattern applies for all use cases onboarded to DEX.

Within a particular JSON file, you will find the custom sender manifest schema for a given use case, as well as file routing configuration.  The sender manifest schema defines validation rules for the fields and values senders provide when uploading a file.  We currently support two versions of schemas, v1 and v2.

#### upload-processor

`/BulkFileUploadFunctionApp` event triggered azure function to process uploaded files.  This component of the Upload API is responsible for normalizing the filename and attached metadata for a given uploaded file, as well as routing the file to the appropriate routing or EDAV storage accuonts.


#### Smoke tests

- `/tests/smoke/kotlin` - The Upload API's official smoke test suite for testing all onboarded use cases across all e
- `/tests/smoke/playwright`

#### spikes

`/spikes` exploratory development spikes, archive.

### TUS Protocol

This repository is using the TUS resumable upload protocol: [https://tus.io/](https://tus.io/),
and reference implementation: [https://github.com/tus/tusd](https://github.com/tus/tusd)

## Example Usage

Example clients, for back-end or browser (front-end), to upload files:
[https://github.com/CDCgov/data-exchange-api-examples](https://github.com/CDCgov/data-exchange-api-examples)

## Future Improvements

- Rework of the tus service to use Tus v2 as a library within a GoLang service
- Simplification of Tus hooks
- Simplification of PS API integration to only focus on reporting
- Improved parallel upload experience with a distributed file lock
- Implementation of Command/Event architecture
- Deployment to Kubernetes instead of Azure App Service

## Public Domain Standard Notice

This repository constitutes a work of the United States Government and is not
subject to domestic copyright protection under 17 USC § 105. This repository is in
the public domain within the United States, and copyright and related rights in
the work worldwide are waived through the [CC0 1.0 Universal public domain dedication](https://creativecommons.org/publicdomain/zero/1.0/).
All contributions to this repository will be released under the CC0 dedication. By
submitting a pull request you are agreeing to comply with this waiver of
copyright interest.

## License Standard Notice

The repository utilizes code licensed under the terms of the Apache Software
License and therefore is licensed under ASL v2 or later.

This source code in this repository is free: you can redistribute it and/or modify it under
the terms of the Apache Software License version 2, or (at your option) any
later version.

This source code in this repository is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the Apache Software License for more details.

You should have received a copy of the Apache Software License along with this
program. If not, see http://www.apache.org/licenses/LICENSE-2.0.html

The source code forked from other open source projects will inherit its license.

## Privacy Standard Notice

This repository contains only non-sensitive, publicly available data and
information. All material and community participation is covered by the
[Disclaimer](https://github.com/CDCgov/template/blob/master/DISCLAIMER.md)
and [Code of Conduct](https://github.com/CDCgov/template/blob/master/code-of-conduct.md).
For more information about CDC's privacy policy, please visit [http://www.cdc.gov/other/privacy.html](https://www.cdc.gov/other/privacy.html).

## Contributing Standard Notice

Anyone is encouraged to contribute to the repository by [forking](https://help.github.com/articles/fork-a-repo)
and submitting a pull request. (If you are new to GitHub, you might start with a
[basic tutorial](https://help.github.com/articles/set-up-git).) By contributing
to this project, you grant a world-wide, royalty-free, perpetual, irrevocable,
non-exclusive, transferable license to all users under the terms of the
[Apache Software License v2](http://www.apache.org/licenses/LICENSE-2.0.html) or
later.

All comments, messages, pull requests, and other submissions received through
CDC including this GitHub page may be subject to applicable federal law, including but not limited to the Federal Records Act, and may be archived. Learn more at [http://www.cdc.gov/other/privacy.html](http://www.cdc.gov/other/privacy.html).

## Records Management Standard Notice

This repository is not a source of government records, but is a copy to increase
collaboration and collaborative potential. All government records will be
published through the [CDC web site](http://www.cdc.gov).

## Additional Standard Notices

Please refer to [CDC's Template Repository](https://github.com/CDCgov/template)
for more information about [contributing to this repository](https://github.com/CDCgov/template/blob/master/CONTRIBUTING.md),
[public domain notices and disclaimers](https://github.com/CDCgov/template/blob/master/DISCLAIMER.md),
and [code of conduct](https://github.com/CDCgov/template/blob/master/code-of-conduct.md).
