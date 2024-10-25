import { readFileSync } from 'fs';
import { resolve } from 'path';

export const SMALL_FILENAME = '10KB-test-file';
export const LARGE_FILENAME = '10MB-test-file';

export const SMALL_FILEPATH = getResourceFilepath(SMALL_FILENAME);
export const LARGE_FILEPATH = getResourceFilepath(LARGE_FILENAME);

export const API_URL = process.env.SERVER_URL ?? 'http://localhost:8080';
export const API_FILE_ENDPOINT = `${API_URL}/files/`;
export const API_INFO_ENDPOINT = `${API_URL}/info/`;

export type Metadata = {
  dataStream: string;
  route: string;
};

export type FileSelection = {
  name: string;
  mimeType: string;
  buffer: Buffer;
};

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

export function getResourceFilepath(filename: string): string {
  return resolve(__dirname, '..', 'resources', filename);
}

export function getFileBuffer(filename: string): Buffer {
  const filepath = getResourceFilepath(filename);
  return readFileSync(filepath);
}

export function getFileJSON(filename: string): any {
  const filepath = getResourceFilepath(filename);
  return JSON.parse(readFileSync(filepath).toString());
}

export function getFileSelection(filename: string): FileSelection {
  const filepath = getResourceFilepath(filename);
  return {
    name: filename,
    mimeType: 'text/plain',
    buffer: getFileBuffer(filepath)
  };
}

export function getMetadataObjects(): Metadata[] {
  const filepath = getResourceFilepath('metadata.json');
  return JSON.parse(readFileSync(filepath).toString());
}

export function getTestCases(filename: string): TestCase[] {
  const filepath = getResourceFilepath(filename);
  return JSON.parse(readFileSync(filepath).toString());
}

export function normalizeValidationErrors(validationErrors: string[] | null | undefined) {
  if (!validationErrors) {
    return [];
  }
  const errorSet = new Set(validationErrors);
  const uniqArray = [...errorSet];
  return uniqArray.filter(item => item != 'validation failure');
}
