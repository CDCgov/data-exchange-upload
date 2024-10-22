import { test } from '@playwright/test';
import dotenv from 'dotenv';
import { resolve } from 'path';
import { v4 as uuidv4 } from 'uuid';
import tusClient from '../tusclient';

dotenv.config({ path: '../../../../upload-server/.env' });

// Use test.describe to group your tests and use hooks like beforeAll
test.describe('File Upload and Trace Response Flow', () => {
  const filename = resolve(__dirname, '..', '..', 'upload-files', '10KB-test-file');
  const apiURL = `${process.env.SERVER_URL ?? 'http://localhost:8080'}/files`;
  const context = tusClient.newContext(apiURL);

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
      ...overrides
    };
  }

  // Parameterize your tests for different metadata scenarios
  const testCases = [
    {
      name: 'Destination ID not provided',
      metadata: createMetadata({ meta_ext_event: 'testevent1' }),
      expectedStatusCode: 400,
      expectedErrorMessage: 'field meta_destination_id was missing',
      pstestProperty: 'data_stream_id',
      psexpectedErrorMessage: 'NOT_PROVIDED'
    },
    {
      name: 'Event type not provided',
      metadata: createMetadata({ meta_destination_id: 'dextesting' }),
      expectedStatusCode: 400,
      expectedErrorMessage: 'field meta_ext_event was missing',
      pstestProperty: 'data_stream_route',
      psexpectedErrorMessage: 'NOT_PROVIDED'
    },
    {
      name: 'Destination ID and event type are mismatched',
      metadata: createMetadata({
        meta_destination_id: 'invalidDestination',
        meta_ext_event: 'invalidEvent'
      }),
      expectedStatusCode: 400,
      expectedErrorMessage:
        'open v1/invaliddestination-invalidevent.json: no such file or directory',
      pstestProperty: 'data_stream_id',
      psexpectedErrorMessage: 'invalidDestination'
    }
  ];

  testCases.forEach(
    ({
      name,
      metadata,
      expectedStatusCode,
      expectedErrorMessage
      // pstestProperty,
      // psexpectedErrorMessage
    }) => {
      test(name, async ({ request }) => {
        const response = await context.upload(filename, metadata);
        response.assertError(expectedStatusCode);
        response.assertValidationErrors(expectedErrorMessage);
        //  const response = await request.get(psApiUrl);
        //   const responseBody = await response.json();
        //   expect(responseBody).toHaveProperty(pstestProperty);
        //   expect(responseBody[pstestProperty]).toBe(psexpectedErrorMessage);
      });
    }
  );

  // Act & Assert
  test('upload and get report response', async ({ request }) => {
    const metadata = createMetadata({
      meta_destination_id: 'dextesting',
      meta_ext_event: 'testevent1'
    });

    const response = await context.upload(filename, metadata);

    // // Assert: Validate the data coming back from PS API
    // expect(responseBody).toHaveProperty('upload_id');
    // expect(responseBody.upload_id).toBe(uploadId); // assuming uploadId matches

    // // Validate reports array is present and has at least one report
    // expect(responseBody).toHaveProperty('reports');
    // expect(Array.isArray(responseBody.reports)).toBeTruthy();
    // expect(responseBody.reports.length).toBeGreaterThan(0);

    // // Validate the structure of the first report if specific validation is needed
    // // e.g., checking if the first report's stage_name is 'dex-metadata-verify'
    // expect(responseBody.reports[0]).toHaveProperty(
    //   'stage_name',
    //   'dex-metadata-verify'
    // );

    // Additional assertions can be added based on the structure of your response
    // For example, checking status code
    response.assertSuccess();
  });
});
