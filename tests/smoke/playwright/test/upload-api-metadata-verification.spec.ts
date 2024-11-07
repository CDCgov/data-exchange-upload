import { expect, test } from '@playwright/test';
import {
  API_FILE_ENDPOINT,
  SMALL_FILEPATH,
  TestCase,
  getTestCases,
  normalizeValidationErrors
} from '../resources/test-utils';
import tusClient from '../tus-playwright';

export type TestSuite = {
  suiteName: string;
  testCases: TestCase[];
};

// Use test.describe to group your tests and use hooks like beforeAll
test.describe(
  'Metadata Validation',
  {
    tag: ['@api', '@metadata']
  },
  () => {
    const filename = SMALL_FILEPATH;
    const context = tusClient.newContext(API_FILE_ENDPOINT);
    const testCaseFiles: TestSuite[] = [
      {
        suiteName: 'Missing Required Fields',
        testCases: getTestCases('invalid_metadata_missing_required_fields.json')
      },
      {
        suiteName: 'Invalid Values',
        testCases: getTestCases('invalid_metadata_invalid_values.json')
      }
    ];

    testCaseFiles.forEach(({ suiteName, testCases }) => {
      test.describe(suiteName, () => {
        testCases.forEach(({ name, metadata, expectedStatusCode, expectedErrorMessages }) => {
          test(name, async () => {
            const response = await context.upload(filename, metadata);
            response.assertError(expectedStatusCode);
            const errors: { [key: string]: any } | null = response.getResponseBodyJson();
            expect(errors).not.toBeNull();
            expect(errors?.validation_errors).not.toBeNull();
            const validationErrors = normalizeValidationErrors(errors?.validation_errors);
            expect(validationErrors.length).toEqual(expectedErrorMessages.length);
            expectedErrorMessages.forEach((error: string) => {
              expect(validationErrors).toContain(error);
            });
          });
        });
      });
    });
  }
);
