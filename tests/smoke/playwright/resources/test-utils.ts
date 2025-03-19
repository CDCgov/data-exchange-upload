import { readFileSync, statSync } from 'fs';
import { resolve } from 'path';

export const SMALL_FILENAME = '10KB-test-file';
export const LARGE_FILENAME = '10MB-test-file';

export const SMALL_FILEPATH = getResourceFilepath(SMALL_FILENAME);
export const LARGE_FILEPATH = getResourceFilepath(LARGE_FILENAME);

export const API_URL = process.env.SERVER_URL ?? 'http://localhost:8080';
export const API_FILE_ENDPOINT = `${API_URL}/files`;
export const API_INFO_ENDPOINT = `${API_URL}/info`;

export const TEST_DATA_STREAM_ID = 'dextesting';
export const TEST_DATA_STREAM_ROUTE = 'testevent1';

export const DATE_TIME_REGEX = /[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}.*[0-9]*Z/;

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

export type Manifest = {
  data_stream_id: string;
  data_stream_route: string;
  received_filename: string;
  sender_id: string;
  data_producer_id: string;
  jurisdiction: string;
};

export type Metadata = {
  manifest: Manifest;
  delivery_targets: string[];
};

export type ManifestResponse = Manifest & {
  dex_ingest_datetime: string;
  upload_id: string;
};

export type UploadTarget = {
  dataStream: string;
  route: string;
};

export type FileInfo = {
  size_bytes: number;
  updated_at: string;
};

export type UploadStatus = {
  status: 'Initiated' | 'In Progress' | 'Complete';
  chunk_received_at: string;
};

export type Delivery = {
  status: string;
  name: string;
  location: string;
  delivered_at: string;
  issues: DeliveryIssue[] | null;
};

export type DeliveryIssue = {
  level: string;
  message: string;
};

export type InfoResponse = {
  manifest: ManifestResponse;
  file_info: FileInfo;
  upload_status: UploadStatus;
  deliveries: Delivery[];
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

export function getFileSizeBytes(filename: string): number {
  const filepath = getResourceFilepath(filename);
  var stats = statSync(filepath);
  return stats.size;
}

export function getUploadTargets(): UploadTarget[] {
  const filepath = getResourceFilepath('upload_targets.json');
  return JSON.parse(readFileSync(filepath).toString());
}

export function getValidMetadata(): Metadata[] {
  const filepath = getResourceFilepath('valid_metadata.json');
  return JSON.parse(readFileSync(filepath).toString());
}

export function getTestConfigV2(metadata: Metadata[] | null): Metadata {
  if (metadata == null) {
    metadata = getValidMetadata();
  }

  return (
    metadata.find(
      config =>
        config.manifest.data_stream_id == TEST_DATA_STREAM_ID &&
        config.manifest.data_stream_route == TEST_DATA_STREAM_ROUTE
    ) ?? {
      manifest: {
        data_stream_id: TEST_DATA_STREAM_ID,
        data_stream_route: TEST_DATA_STREAM_ROUTE,
        received_filename: 'dex-smoke-test',
        sender_id: 'test sender',
        data_producer_id: 'test-producer-id',
        jurisdiction: 'test-jurisdiction'
      },
      delivery_targets: ['edav', 'ncird']
    }
  );
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
