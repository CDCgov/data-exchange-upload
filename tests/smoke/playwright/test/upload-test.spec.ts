import { expect, test } from '@playwright/test';
import dotenv from "dotenv";
import path from 'path';
import { v4 as uuidv4 } from 'uuid';
import { UploadClient } from '../upload-client';
dotenv.config({ path: '../../../../upload-server/.env' });


// Use test.describe to group your tests and use hooks like beforeAll
test.describe.skip('File Upload and Trace Response Flow', () => {
  let uploadId;
  let accessToken: string;
  let psApiUrl: string;

  const username = process.env.SAMS_USERNAME;
  const password = process.env.SAMS_PASSWORD;
  const url = process.env.UPLOAD_URL;
  const PS_API_URL = process.env.PS_API_URL;
  const fileName = path.resolve(__dirname, '..', '..', 'upload-files', '10KB-test-file');
  const client = new UploadClient(url);

  // Helper function to create metadata with different scenarios
  function createMetadata(overrides = {}) {
    return {
      filename: '10KB-test-file',
      filetype: 'text/plain',
      meta_ext_source: 'IZGW',
      meta_ext_sourceversion: 'V2022-12-31',
      meta_ext_entity: 'DD2',
      meta_username: 'ygj6@cdc.gov',
      meta_ext_objectkey: uuidv4(),
      meta_ext_filename: '10KB-test-file',
      meta_ext_submissionperiod: '1',
      ...overrides,
    };
  }

  function createPSEndPoint() {
    psApiUrl = `${PS_API_URL}/api/report/uploadId/${uploadId}`;
    return psApiUrl;
  }

  function assertErrorResponse(error, expectedStatusCode, expectedErrorMessageSubstring) {
    const errorMessage = error instanceof Error ? error.message : String(error);

    // Define regex for extracting the HTTP status code 
    const statusCodeRegex = /response code: (\d+)/;

    // Use RegExp.exec() to extract the HTTP status code
    const statusCodeMatch = statusCodeRegex.exec(errorMessage);
    const statusCode = statusCodeMatch ? parseInt(statusCodeMatch[1], 10) : undefined;

    // Assert that the extracted status code matches the expected status code
    expect(statusCode).toBe(expectedStatusCode);

    // Further, assert that the error message contains the specific error detail
    expect(errorMessage).toContain(expectedErrorMessageSubstring);

  }

  // New function to extract upload ID from an error message
  function extractUploadIdFromError(error) {
    const errorMessage = error instanceof Error ? error.message : String(error);
    const uploadIdRegex = /"upload_id": "([\w-]+)"/;
    const uploadIdMatch = uploadIdRegex.exec(errorMessage);
    return uploadIdMatch ? uploadIdMatch[1] : null;
  }


  // Arrange
  // Correctly use test.beforeAll at the describe level to perform setup actions
  test.beforeAll(async ({ request }) => {
    accessToken = await client.login(username, password);
    if (!accessToken) throw new Error("Login failed.");
  });

  // Parameterize your tests for different metadata scenarios
  const testCases = [
    {
      name: 'Destination ID not provided',
      metadata: createMetadata({ meta_ext_event: 'testevent1' }),
      expectedStatusCode: 500,
      expectedErrorMessage: "Missing one or more required metadata fields: ['meta_destination_id']",
      pstestProperty: 'data_stream_id',
      psexpectedErrorMessage: 'NOT_PROVIDED',
    },
    {
      name: 'Event type not provided',
      metadata: createMetadata({ meta_destination_id: 'dextesting' }),
      expectedStatusCode: 500,
      expectedErrorMessage: "Missing one or more required metadata fields: ['meta_ext_event']",
      pstestProperty: 'data_stream_route',
      psexpectedErrorMessage: 'NOT_PROVIDED',
    },
    {
      name: 'Destination ID and event type are mismatched',
      metadata: createMetadata({ meta_destination_id: 'invalidDestination', meta_ext_event: 'invalidEvent' }),
      expectedStatusCode: 500,
      expectedErrorMessage: 'Not a recognized combination of meta_destination_id (invalidDestination) and meta_ext_event (invalidEvent)',
      pstestProperty: 'data_stream_id',
      psexpectedErrorMessage: 'invalidDestination',
    },
    {
      name: 'Config schema definition invalid',
      metadata: createMetadata({ meta_destination_id: 'dextesting', meta_ext_event: 'testevent1', invalidField: 'thisShouldNotBeHere' }),
      expectedStatusCode: 500,
      expectedErrorMessage: 'Config schema definition is invalid',
      pstestProperty: 'data_stream_id',
      psexpectedErrorMessage: 'invalidDestination',
    },
  ];


  testCases.forEach(({ name, metadata, expectedStatusCode, expectedErrorMessage, pstestProperty, psexpectedErrorMessage }) => {
    test(name, async ({ request }) => {

      try {
        uploadId = await client.uploadFileAndGetId(accessToken, fileName, metadata);

      } catch (error) {
        assertErrorResponse(error, expectedStatusCode, expectedErrorMessage);
        uploadId = extractUploadIdFromError(error);
        const psApiUrl = `${PS_API_URL}/api/report/uploadId/${uploadId}`;
        const response = await request.get(psApiUrl);
        const responseBody = await response.json();
        expect(responseBody).toHaveProperty(pstestProperty);
        expect(responseBody[pstestProperty]).toBe(psexpectedErrorMessage);

      }

    });
  });

  // Act & Assert 
  test('upload and get report response', async ({ request }) => {

    const metadata = createMetadata();
    metadata['meta_destination_id'] = 'dextesting';
    metadata['meta_ext_event'] = 'testevent1';
    uploadId = await client.uploadFileAndGetId(accessToken, fileName, metadata);
    psApiUrl = createPSEndPoint();

    // Act: Query the PS API with the obtained uploadId and accessToken
    const response = await request.get(psApiUrl);

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
