import { expect, test } from '@playwright/test';
import dotenv from 'dotenv';
import { readFileSync } from 'fs';
import { resolve } from 'path';
import tusClient from '../tus-playwright';

dotenv.config({ path: '../../../../upload-server/.env' });

export type TestSuite = {
  suiteName: string;
  testCaseFilename: string;
};

export type TestCase = {
  name: string;
  metadata: { [key: string]: string };
  expectedStatusCode: number;
  expectedErrorMessages: string[];
};

const normalizeErrors = (validationErrors: string[] | null | undefined) => {
  if (!validationErrors) {
    return [];
  }
  const errorSet = new Set(validationErrors);
  const uniqArray = [...errorSet];
  return uniqArray.filter(item => item != 'validation failure');
};

// Use test.describe to group your tests and use hooks like beforeAll
test.describe(
  'Metadata Validation',
  {
    tag: ['@api', '@metadata']
  },
  () => {
    const filename = resolve(__dirname, '..', 'resources', '10KB-test-file');
    const apiURL = `${process.env.SERVER_URL ?? 'http://localhost:8080'}/files`;
    const context = tusClient.newContext(apiURL);
    const testCaseFiles: TestSuite[] = [
      {
        suiteName: 'Missing Required Fields',
        testCaseFilename: resolve(
          __dirname,
          '..',
          'resources',
          'invalid_metadata_missing_required_fields.json'
        )
      },
      {
        suiteName: 'Invalid Values',
        testCaseFilename: resolve(
          __dirname,
          '..',
          'resources',
          'invalid_metadata_invalid_values.json'
        )
      }
    ];

    testCaseFiles.forEach(({ suiteName, testCaseFilename }) => {
      test.describe(suiteName, () => {
        const testCases: TestCase[] = JSON.parse(readFileSync(testCaseFilename).toString());

        testCases.forEach(({ name, metadata, expectedStatusCode, expectedErrorMessages }) => {
          test(name, async () => {
            const response = await context.upload(filename, metadata);
            response.assertError(expectedStatusCode);
            const errors: { [key: string]: any } | null = response.getResponseBodyJson();
            expect(errors).not.toBeNull();
            expect(errors?.validation_errors).not.toBeNull();
            const validationErrors = normalizeErrors(errors?.validation_errors);
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
