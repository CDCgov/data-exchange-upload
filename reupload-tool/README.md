# Upload API File Re-Upload Tool

This program takes a file that was previously uploaded, and re-uploads it from a specified environment.  This is useful
whenever a file did not get processed by downstream programs properly, and needs to be re-uploaded to re-trigger the downstream
pipeline.  The re-uploaded file's metadata remains intact and unchanged.  However, this tool does offer the ability to
set a new name for the re-uploaded file.

## Usage

### Setup

Install necessary software dependencies:
1. Java JVM 15 or greater
2. Gradle (using gradlew)

### Input CSV

To specify one or more files to be re-uploaded, generate a CSV file with the following columns:

- src - The name of the file to be re-uploaded, including any path prefixes.
- dest - The name of the file to be re-uploaded under.  Can be the same name as `src` without the path prefixes.
- srcaccountid - The unique label or identifier of the storage account where the `src` file is located.  Available options are `edav` and `routing`.

### Build and Run

This program uses Gradle to build and run a jar file.  You can do this with the following command:

`gradle run -Dcsv=input.csv` where `input.csv` is your CSV file generated above.