# Upload API Smoke Tests with TestNG

This is a Kotlin/Gradle project that uses the TestNG framework to automate smoke testing for the Upload API. These tests
upload actual files and verify the functionality for metadata verification, file copy, and Processing Status API
integration. They are intended to be run after a release to the tst, stg, or prd environments. They can also be run
locally on a machine within the CDC network.

## Local Setup

The following tools need to be installed on your machine:

- Java JDK 17
- Gradle

### Install Gradle Dependencies

Next, run `gradle build` to install dependencies for this project. This also installs the TestNG dependency.  **Note
that you may need to turn off Zscalar in order for this operation to be successful.**

### Environment Setup

Next, set required environment variables. This can be done by setting local gradle properties in a `local.properties`
file at the root level, or passing them in on the command line as Java system properties. To see a list of required
variables, look at the `src/test/kotlin/util/EnvConfig.kt` file.
These environment variables are how you target different DEX environments. For example, for the dev environment, all
environment variables need to point to URLs, endpoints, and services uses by the Upload API dev environment.

### Running tests

This project contains a set of test suites that define the tests to be run. These suites are grouped by environment and
are broken up by use case.
They are run by executing the `gradle test` comment with a few gradle properties that are passed in as command line
arguments:

- `manifestFilter` - This is a required argument that allows you to select a subset of use cases you want to run the
  tests against. It is a comma-separated list of values for each key that you want to run. Available manifests are added
  to a file which is under resources folder in json format. By default, the tests will run for all use cases.

#### Examples:

- Run all use cases:

`gradle test`

## Data Providers for TestNG

This project includes a `DataProvider` utility class that supplies test data to TestNG tests. The class uses the Jackson
library to read and filter JSON manifests based on criteria specified through system properties.

### How It Works

1. Loading Manifests:

- The `DataProvider` class reads JSON manifest files specified in the data provider methods.

2. Filtering Manifests:

- A system property `manifestFilter` can be set to define filtering criteria.
- This filter is a semicolon-separated string of key-value pairs.
- Each key can have multiple comma-separated values.

### Example Usage

To filter specific key-value pairs in the JSON manifests, use the `manifestFilter` system property.
Here are the example commands to run the tests with manifest filters:

`gradle test -PmanifestFilter='meta_destination_id=ndlp&meta_ext_source=IZGW'`
`gradle test -PmanifestFilter='jurisdiction=AKA,CA;data_stream_route=csv,other&sender_id=CA-ABCs,IZGW'`

We can run tests for specified manifest in a single command line argument.

### Future Improvements

- Parallelization for data provided test cases.





