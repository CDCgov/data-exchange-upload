import { expect, test } from '@playwright/test';
import dotenv from "dotenv";
import path from 'path';
import { v4 as uuidv4 } from 'uuid';
import { loginAndGetToken } from '../login';
import { uploadFileAndGetId } from '../upload';
dotenv.config({ path: "../../../.env" });

// Use test.describe to group your tests and use hooks like beforeAll
test.describe('File Upload and Trace Response Flow', () => {
  let uploadId;
  let accessToken: string;
  let psApiUrl: string;

  // Create metadata object
  const metadata = {
    filename: "10MB-test-file",
    filetype: "text/plain",
    meta_destination_id: "dextesting",
    meta_ext_event: "testevent1",
    meta_ext_source: "IZGW",
    meta_ext_sourceversion: "V2022-12-31",
    meta_ext_entity: "DD2",
    meta_username: "ygj6@cdc.gov",
    meta_ext_objectkey: uuidv4(),
    meta_ext_filename: "10MB-test-file",
    meta_ext_submissionperiod: '1',
  };



  // Arrange
  // Correctly use test.beforeAll at the describe level to perform setup actions
  test.beforeAll(async ({ request }) => {
    try {
      const url = process.env.UPLOAD_URL;
      const PS_API_URL = process.env.PS_API_URL;
      const fileName = path.resolve(__dirname, '..', '..', 'upload-files', '10MB-test-file');
      accessToken = await loginAndGetToken();
      uploadId = await uploadFileAndGetId(accessToken, fileName, url, metadata);
      psApiUrl = `${PS_API_URL}/api/report/uploadId/${uploadId}`;

    } catch (error) {
      console.error("Setup failed:", error);
      throw new Error("Failed to complete test setup.");
    }
  });

  // Act & Assert 
  test('upload and get trace response', async ({ request }) => {

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
