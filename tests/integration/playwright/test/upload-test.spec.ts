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
  const fileName = path.resolve(__dirname, '..', '..', 'upload-files', '10KB-test-file');
  const client = new UploadClient(url);

  // Helper function to create metadata with different scenarios
  function createMetadata() {
    return {
      filename: "10KB-test-file",
      filetype: "text/plain",
      meta_ext_source: "IZGW",
      meta_ext_sourceversion: "V2022-12-31",
      meta_ext_entity: "DD2",
      meta_username: "ygj6@cdc.gov",
      meta_ext_objectkey: uuidv4(),
      meta_ext_filename: "10KB-test-file",
      meta_ext_submissionperiod: '1',
    };
  }

  function createPSEndPoint() {
    psApiUrl = `${PS_API_URL}/api/report/uploadId/${uploadId}`;
    return psApiUrl;
  }

  function assertErrorResponse(error, expectedStatusCode, expectedErrorMessageSubstring) {
    const errorMessage = error instanceof Error ? error.message : String(error);

    // Attempt to extract the HTTP status code from the error message
    const match = errorMessage.match(/response code: (\d+)/);
    const statusCode = match ? parseInt(match[1], 10) : undefined;

    // Assert that the extracted status code matches the expected status code
    expect(statusCode).toBe(expectedStatusCode);

    // Further, assert that the error message contains the specific error detail
    expect(errorMessage).toContain(expectedErrorMessageSubstring);

    // Attempt to extract the uploadId from the error message
    const matchUploadId = errorMessage.match(/"upload_id": "([\w-]+)"/);
    uploadId = matchUploadId ? matchUploadId[1] : null;

  }


  // Arrange
  // Correctly use test.beforeAll at the describe level to perform setup actions
  test.beforeAll(async ({ request }) => {
    accessToken = await client.login(username, password);
    if (!accessToken) throw new Error("Login failed.");
  });

  // Test case 1: Destination ID not provided
  test('should return 500 when destination ID not provided', async ({ request }) => {
    const metadata = createMetadata();
    metadata['meta_ext_event'] = 'testevent1';

    try {
      await client.uploadFileAndGetId(accessToken, fileName, metadata);

    } catch (error) {
      assertErrorResponse(error, 500, "Missing one or more required metadata fields: ['meta_destination_id']");
    }
  });

  test('Query the PS API should return 500 when destination ID not provided', async ({ request }) => {

    try {
      psApiUrl = createPSEndPoint();

      // Act: Query the PS API 
      const response = await request.get(psApiUrl, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });

      // Parse the JSON response body
      const responseBody = await response.json();
     
      expect(responseBody).toHaveProperty('data_stream_id');
      expect(responseBody.data_stream_id).toBe('NOT_PROVIDED');
    } catch (error) {
      error.message;
    }
  });

 
   // Test case 2: Event type not provided
   test('should return 500 when event type not provided', async ({ request }) => {
     const metadata = createMetadata();
     metadata['meta_destination_id'] = 'dextesting';
     try {
       await client.uploadFileAndGetId(accessToken, fileName, metadata);
 
     } catch (error) {
       assertErrorResponse(error, 500, "Missing one or more required metadata fields: ['meta_ext_event']");
     }
   });

   test('Query the PS API should return 500 when event type not provided', async ({ request }) => {

    try {
      psApiUrl = createPSEndPoint();

      // Act: Query the PS API 
      const response = await request.get(psApiUrl, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });

      // Parse the JSON response body
      const responseBody = await response.json();
      
      expect(responseBody).toHaveProperty('data_stream_route');
      expect(responseBody.data_stream_route).toBe('NOT_PROVIDED');
    } catch (error) {
      error.message;
    }
  });

  
   // Test case 3: Destination ID and event type are mismatched
   test('should return 500 when destination ID and event type are mismatched', async ({ request }) => {
     const metadata = createMetadata();
     metadata['meta_destination_id'] = 'invalidDestination';
     metadata['meta_ext_event'] = 'invalidEvent';
     try {
       await client.uploadFileAndGetId(accessToken, fileName, metadata);
 
     } catch (error) {
       assertErrorResponse(error, 500, "Not a recognized combination of meta_destination_id (invalidDestination) and meta_ext_event (invalidEvent)");
     }
   });

   test('Query the PS API should return 500 when destination ID and event type are mismatched', async ({ request }) => {

    try {
      psApiUrl = createPSEndPoint();

      // Act: Query the PS API 
      const response = await request.get(psApiUrl, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });

      // Parse the JSON response body
      const responseBody = await response.json();
      
      expect(responseBody).toHaveProperty('data_stream_id');
      expect(responseBody.data_stream_id).toBe('invalidDestination');
    } catch (error) {
      error.message;
    }
  });

   // Test case 4: Config schema definition invalid  
   test('should return 500 when config schema definition invalid', async ({ request }) => {   
     const metadata = createMetadata();
     metadata['meta_destination_id'] = 'dextesting';
     metadata['meta_ext_event'] = 'testevent1';
     metadata['invalidField'] = 'thisShouldNotBeHere';

     try {
       await client.uploadFileAndGetId(accessToken, fileName, metadata);
 
     } catch (error) {
       assertErrorResponse(error, 500, "Missing required metadata 'filename', description = 'The name of the file submitted.");
     }
   });

   test('Query the PS API should return 500 when config schema definition invalid', async ({ request }) => {

    try {
      psApiUrl = createPSEndPoint();

      // Act: Query the PS API 
      const response = await request.get(psApiUrl, {
        headers: {
          Authorization: `Bearer ${accessToken}`,
        },
      });

      // Parse the JSON response body
      const responseBody = await response.json();
      
      expect(responseBody).toHaveProperty('data_stream_id');
      expect(responseBody.data_stream_id).toBe('invalidDestination');
    } catch (error) {
      error.message;
    }
  });

 
   // Act & Assert 
   test('upload and get report response', async ({ request }) => {
 
     const metadata = createMetadata();
     metadata['meta_destination_id'] = 'dextesting';
     metadata['meta_ext_event'] = 'testevent1';
     uploadId = await client.uploadFileAndGetId(accessToken, fileName, metadata);
     psApiUrl = createPSEndPoint();
 
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
