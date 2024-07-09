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

We can run tests with single use case or multiple use cases specified in a single command line argument.

- Run only the test use case:

   A single use case can accept one key-value pair and/or more than one key-value pair, which are separated by a comma (,). Each key-value pair is split by (:) to separate keys from values.

Example: `gradle test -PuseCases=data_stream_id:pulsenet`
Example: `gradle test -PuseCases=data_stream_id:influenza-vaccination,data_stream_route:csv`

- Run a select few use cases:

  Multiple use cases are separated by a semicolon (;) and the use cases should be wrapped in quotation marks. Each use case has more than one key-value pair, which is separated by a comma (,). Each key-value pair is split by (:) to separate keys from values.

Example: `gradle test -PuseCases="data_stream_id:eicr;data_stream_route:hl7v2"`
Example: `gradle test -PuseCases="data_stream_id:abcs,data_stream_route:csv;data_stream_id:ed3n,data_stream_route:other"`

## Future Improvements

- Parallelization for data provided test cases.



