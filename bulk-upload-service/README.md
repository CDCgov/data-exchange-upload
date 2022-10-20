# Data Exchange - Bulk Upload Service
The bulk upload service is a RESTful web service whose sole purpose is to facilitate bulk file uploads.
The service will support several types of uploads including multipart file uploads as well as resumable file uploads.

The bulk upload service is a Spring Boot application written in Kotlin.  It is intended to be deployed as an Azure App Service.

## Build and Deploy from CLI
1. Azure login with your **SU account**.  Otherwise, you will not have permission to deploy to the Azure subscription.
    ```bash
    az login
   ```
2. Build the project
    ```bash
    ./gradlew clean build -x test
    ```
3. Deployment
    ```bash
    ./gradlew azureWebAppDeploy
    ```