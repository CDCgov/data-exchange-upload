# Upload API Smoke Tests with TestNG

This is a Kotlin/Gradle project that uses the TestNG framework to automate smoke testing for the Upload API.  These tests upload actual files and verify the functionality for metadata verification, file copy, and Processing Status API integration.  They are intended to be run after a release to the tst, stg, or prd environments.  They can also be run locally on a machine within the CDC network.  

## Local Setup

The following tools need to be installed on your machine:
- Java JDK 11
- Gradle

### Install Gradle Dependencies

Next, run `gradle build` to install dependencies for this project.  This also installs the TestNG dependency.  **Note that you may need to turn off Zscalar in order for this operation to be successful.**

### Environment Setup

Next, set required environment variables.  This can be done by setting local gradle properties in a `local.properties` file at the root level, or passing them in on the command line as Java system properties.  To see a list of required variables, look at the `src/test/kotlin/util/EnvConfig.kt` file.
These environment variables are how you target different DEX environments.  For example, for the dev environment, all environment variables need to point to URLs, endpoints, and services uses by the Upload API dev environment.

### Running tests

This project contains a set of test suites that define the tests to be run.  These suites are grouped by environment and are broken up by use case.
They are run by executing the `gradle test` comment with a few gradle properties that are passed in as command line arguments:

- `useCases` - This is an optional argument that allows you to select a subset of use cases you want to run the tests against.  It is a comma-separated list of use cases that you want to run.  Available use cases are defined by the name of the XML files within the environment folders.  By default, the tests will run for all use cases.

#### Examples:

- Run all use cases:

`gradle test`

- Run only the test use case:

`gradle test -PuseCases=dextesting-testevent1`

- Run a select few use cases:

`gradle test -PuseCases=aims-celr-csv,aims-celr-hl7`

## Future Improvements

- Parallelization for data provided test cases.