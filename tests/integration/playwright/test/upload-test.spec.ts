import { expect, test } from '@playwright/test';
import { getTraceResponse, start } from '../index';

// Use test.describe to group your tests and use hooks like beforeAll
test.describe('File Upload and Trace Response Flow', () => {
  let uploadId: string;
  let accessToken: string;

  // Arrange
  // Correctly use test.beforeAll at the describe level to perform setup actions
  test.beforeAll(async () => {
    // Perform setup, including logging in and uploading the file to get an upload ID
    const result = await start();
    uploadId = result.uploadId;
    accessToken = result.accessToken;
  });

  // Act & Assert 
  test('upload and get trace response', async () => {
    // Assert that we have an upload ID and an access Token
    expect(uploadId).toBeTruthy();
    expect(accessToken).toBeTruthy();

    // Act: Query the PS API with the obtained uploadId and accessToken
    const traceResponse = await getTraceResponse(uploadId, accessToken);

    // Assert: Validate the data coming back from PS API equals an expected value    
    expect(traceResponse).toHaveProperty('expectedProperty');
    
  });
});
