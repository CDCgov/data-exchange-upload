import { expect, test } from '@playwright/test';
import dotenv from "dotenv";
import path from 'path';
import { v4 as uuidv4 } from 'uuid';
import { UploadClient } from '../upload-client';
dotenv.config({ path: "../../.env" });

// Use test.describe to group your tests and use hooks like beforeAll
test.describe('File Upload and Trace Response Flow', () => {
  let uploadId;
  let accessToken: string;
  let psApiUrl: string;

  const username = process.env.SAMS_USERNAME;
  const password = process.env.SAMS_PASSWORD;
  const url = process.env.UPLOAD_URL;
  const PS_API_URL = process.env.PS_API_URL;
  const fileName = path.resolve(__dirname, '..', '..', 'upload-files', '10MB-test-file');
  const client = new UploadClient(url);

  // Helper function to create metadata with different scenarios
  function createMetadata({ destinationId, eventType }) {
    return {
      filename: "10MB-test-file",
      filetype: "text/plain",
      meta_destination_id: destinationId,
      meta_ext_event: eventType,
      meta_ext_source: "IZGW",
      meta_ext_sourceversion: "V2022-12-31",
      meta_ext_entity: "DD2",
      meta_username: "ygj6@cdc.gov",
      meta_ext_objectkey: uuidv4(),
      meta_ext_filename: "10MB-test-file",
      meta_ext_submissionperiod: '1',
    };
  }

  // Arrange
  // Correctly use test.beforeAll at the describe level to perform setup actions
  test.beforeAll(async ({ request }) => {
    accessToken = await client.login(username, password);
    if (!accessToken) throw new Error("Login failed.");
  });

  // Test case 1: Destination ID not provided
  test('should return 500 when destination ID not provided', async ({ request }) => {
    const metadata = createMetadata({ destinationId: undefined, eventType: "testevent1" });

    try {
      await client.uploadFileAndGetId(accessToken, fileName, metadata);
      
    } catch (error) {
      // Convert error to string to ensure matching works
      const errorMessage = error instanceof Error ? error.message : String(error);

      // Attempt to extract the HTTP status code from the error message
      const match = errorMessage.match(/response code: (\d+)/);
      const statusCode = match ? parseInt(match[1], 10) : undefined;

      // Assert that the extracted status code is 500
      expect(statusCode).toBe(500);
      // Further, assert that the error message contains the specific error detail
      expect(errorMessage).toContain("Not a recognized combination of meta_destination_id (undefined) and meta_ext_event (testevent1)");
    }
  });

  // Test case 2: Event type not provided
  test('should return 500 when event type not provided', async ({ request }) => {
    const metadata = createMetadata({ destinationId: "dextesting", eventType: undefined });
    try {
      const response = await client.uploadFileAndGetId(accessToken, fileName, metadata);

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);

      // Attempt to extract the HTTP status code from the error message
      const match = errorMessage.match(/response code: (\d+)/);
      const statusCode = match ? parseInt(match[1], 10) : undefined;

      // Assert that the extracted status code is 500
      expect(statusCode).toBe(500);
      // Further, assert that the error message contains the specific error detail
      expect(errorMessage).toContain("Not a recognized combination of meta_destination_id (dextesting) and meta_ext_event (undefined)");
    }
  });

  // Test case 3: Destination ID and event type are mismatched
  test('should return 500 when destination ID and event type are mismatched', async ({ request }) => {
    const metadata = createMetadata({ destinationId: "invalidDestination", eventType: "invalidEvent" });
    try {
      const response = await client.uploadFileAndGetId(accessToken, fileName, metadata);

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);

      // Attempt to extract the HTTP status code from the error message
      const match = errorMessage.match(/response code: (\d+)/);
      const statusCode = match ? parseInt(match[1], 10) : undefined;

      // Assert that the extracted status code is 500
      expect(statusCode).toBe(500);
      // Further, assert that the error message contains the specific error detail
      expect(errorMessage).toContain("Not a recognized combination of meta_destination_id (invalidDestination) and meta_ext_event (invalidEvent)");
    }
  });

  // Test case 4: Config schema definition invalid  
  test('should return 500 when config schema definition invalid', async ({ request }) => {
    const metadata = {
      meta_destination_id: "dextesting",
      meta_ext_event: "testevent1",
      invalidField: "thisShouldNotBeHere", // Adding an unexpected field
    };
    try {
      const response = await client.uploadFileAndGetId(accessToken, fileName, metadata);

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : String(error);

      // Attempt to extract the HTTP status code from the error message
      const match = errorMessage.match(/response code: (\d+)/);
      const statusCode = match ? parseInt(match[1], 10) : undefined;

      // Assert that the extracted status code is 500
      expect(statusCode).toBe(500);

    }
  });

  // Act & Assert 
  test('upload and get trace response', async ({ request }) => {

    const metadata = createMetadata({ destinationId: "dextesting", eventType: "testevent1" });
    uploadId = await client.uploadFileAndGetId(accessToken, fileName, metadata);
    psApiUrl = `${PS_API_URL}/api/report/uploadId/${uploadId}`;

    // Act: Query the PS API with the obtained uploadId and accessToken
    const response = await request.get(psApiUrl, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    });

    // Parse the JSON response body
    const responseBody = await response.json();

    // Assert: Validate the data coming back from PS API
    expect(responseBody).toHaveProperty('upload_id');
    expect(responseBody.upload_id).toBe(uploadId); // assuming uploadId matches

    // Validate reports array is present and has at least one report
    expect(responseBody).toHaveProperty('reports');
    expect(Array.isArray(responseBody.reports)).toBeTruthy();
    expect(responseBody.reports.length).toBeGreaterThan(0);

    // Validate the structure of the first report if specific validation is needed
    // e.g., checking if the first report's stage_name is 'dex-metadata-verify'
    expect(responseBody.reports[0]).toHaveProperty('stage_name', 'dex-metadata-verify');

    // Additional assertions can be added based on the structure of your response
    // For example, checking status code
    expect(response.status()).toBe(200);

  });
});
