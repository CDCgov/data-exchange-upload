# Upload API End to End Tests with TestNG
This is a Kotlin/Gradle project that uses the TestNG framework to automate end to end testing for the Upload API. These tests upload actual files and verify the functionality for metadata verification, file copy, Processing Status API integration and healthcheck testing. They are intended to be run after a release to the `DEV`, `TEST` or `STAGE` environments or for frequent checking of functionality against any environment, including local environments.

## Local Setup

The following tools need to be installed on your machine:

- [Java JDK 17](https://www.oracle.com/java/technologies/javase/jdk17-archive-downloads.html)
- [Gradle](https://gradle.org/install/)

> [!TIP]
> For Windows users, the gradlew batch file in the repo can be used to execute gradle commands as long as Gradle is installed.

### Install Gradle Dependencies

Next, run `gradle build` to install dependencies for this project. This also installs the TestNG dependency and other related dependencies for the project.  

> [!NOTE]
> You may need to turn off Zscalar in order for this operation to be successful.

### Environment Setup

Copy `local.properites-example` to `local.properties` for the basic required environment variables at the root level of the project. 

The following are the currently configured environment variables that can be set for the tests to run.

| Environment Variable | Required? | Description                                      |
| ---------------------| --------- | -------------------------------------------------|
| environment          | Yes       | Environment to target (LOCAL, DEV, TEST, STAGE)
| upload.url           | Yes       | URL of the upload API
| ps.api.url           | Yes       | URL for Processing Status API  
| sams.username        | No        | SAMS username for authentication (if needed)
| sams.password        | No        | SAMS password for authentication (if needed)

The EnvConfig class (`src/test/kotlin/util/EnvConfig.kt`) reads configuration values from a local.properties file. This setup allows us to manage environment-specific settings, like URLs and credentials.

### Running tests

This project contains a set of test suites that define the tests to be run for different functional areas.  Tests are located in `/src/test/kotlin`.  The test files are:
| Test File      | Purpose        |
|----------------|----------------|
| `FileCopy.kt ` | Testing that the upload api can accept files for different manifest configurations and can upload and transfer files as expected.
| `Health.kt`    | Test the healthcheck endpoint
| `Info.kt `     | Test the /info endpoint 
| `ProcStat.kt`  | Testing processing status reports after an upload

They are run by executing the `gradle test` with the option of using some command line parameters.

#### Optional Parameters

- `manifestFilter` - This argument allows you to select a subset of use cases you want to run the tests against. It is a comma-separated list of values for each key that you want to run. Available manifests are added to a file which is under resources folder in json format. By default, the tests will run for all use cases.

#### Examples:
The following are some examples of test run commands.

> [!TIP]
> The `--tests` parameter lets you select only a specific test file to run or a specific test within a file.

> [!TIP]
> The --rerun command may be needed in order to force tests to rerun, otherwise tests may be skipped if there are no changes.

##### Run All Tests

```
gradle test
```

##### Run tests from a specific file

```
gradle test --tests "FileCopy"
```

##### Run tests from a specific file and a specific test

```
gradle test --tests "FileCopy.shouldUploadFile"
```

##### Run tests filtered by manifest

> [!TIP]
> This filter is a semicolon-separated string of key-value pairs.
> Each key can have multiple comma-separated values.
```
gradle test -PmanifestFilter='data_stream_id=ehdi'
```
```
gradle test -PmanifestFilter='meta_destination_id=ndlp&meta_ext_source=IZGW'
```
```
gradle test -PmanifestFilter='jurisdiction=AKA,CA;data_stream_route=csv,other&sender_id=CA-ABCs,IZGW'
```

## Data Providers for TestNG

This project includes a `DataProvider` utility class that supplies test data to TestNG tests.  The test data being used in tests are essentially test cases that define how a single test can be repeated and validated for different cases defined within the json being used as a source for the DataProvider.

### Data Provider Definitions

The `dataProvider` decorator before a test defines what data provider should be used to pass in data into the test.  These are the currently defined Data Providers in `/src/test/kotlin/util/DataProvider.kt`

| Data Provider Name                      | Associated json file                      | Description                 |
|-----------------------------------------|-------------------------------------------|-----------------------------|
| `versionProvider`                       | N/A                                       | Returns an array of `["v1", "v2"]` for versions
| `validManifestAllProvider`              | `valid_manifests_v2.json`                 | All v2 manifests and path configs |
| `validManifestV1Provider`               | `valid_manifests_v1.json`                 | All v1 manifests and configs | 
| `invalidManifestRequiredFieldsProvider` | `invalid_manifests_required_fields.json`  | Manifests with invalid values |
| `invalidManifestInvalidValueProvider`   | `invalid_manifests_invalid_value.json`    | Manifests with invalid fields  |


### Future Improvements

- Parallelization for data provided test cases.
