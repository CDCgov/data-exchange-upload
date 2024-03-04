import { expect, test } from '@playwright/test';
import config from '../config';
import { loginAndGetToken } from '../login';
import { uploadFileAndGetId } from '../upload';

// Use test.describe to group your tests and use hooks like beforeAll
test.describe('File Upload and Trace Response Flow', () => {
  let uploadId;
  let accessToken: string;
  let psApiUrl: string;

  // Arrange
  // Correctly use test.beforeAll at the describe level to perform setup actions
  test.beforeAll(async ({ request }) => {
    try {
      accessToken = await loginAndGetToken();     
      uploadId = await uploadFileAndGetId(accessToken);      
      psApiUrl = `${config.validateEnv()[3]}/api/trace/uploadId/${uploadId}`;
      
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
    console.error("responseBody:", responseBody);

    // Assert: Validate the data coming back from PS API equals an expected value
    // Replace 'expectedProperty' and 'expectedValue' with actual data you expect
    expect(responseBody).toHaveProperty('expectedProperty');
    expect(responseBody.expectedProperty).toBe('expectedValue');

    // Additional assertions can be added based on the structure of your response
    // For example, checking status code
    expect(response.status).toBe(200);
    
  });
});
