## Overview

This project acts as a java client to send files to CDC's Data Exchange Upload API.

You’ll generate a Java application that follows Gradle’s conventions.

## Prerequisites

- A text editor or IDE
- A Java Development Kit (JDK) >= Version 8
- The latest [Grade distribution](https://gradle.org/install/)

## Setting Up Credentials

This tool must have the following environment variables or create a env.properties file under resources folder with this information.

```bash
USERNAME=<sams_sys_account_name> \
PASSWORD=<sams_sys_account_pswd> \
URL=<dex_staging_url> \
```

## Running the code

- Bundle the application

```bash
./gradle clean jar
```

- Runnign the jar file sample command

```bash
java -DUSERNAME="username" -DPASSWORD="password" -DURL="https://apidev.cdc.gov" -DSMOKE -jar .\dex-upload-client-1.0-SNAPSHOT.jar
```

```bash
java -DUSERNAME="username" -DPASSWORD="password" -DURL="https://apidev.cdc.gov" -DREGRESSION -DConfigsFolder="folder-location" -jar .\dex-upload-client-1.0-SNAPSHOT.jar
```


