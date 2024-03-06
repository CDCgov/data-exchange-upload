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

### Run the tests

Finally, run all tests with `gradle test`.  **This make take several minutes!**.

## Future Improvements

- Paramaterization of file copy tests for each and every use case.
- Organization of tests by use case using better group names or with a `testng.xml` configuration.