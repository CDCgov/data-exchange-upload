import { readFileSync, statSync } from 'fs';
import { resolve } from 'path';

export const SMALL_FILENAME = '10KB-test-file';
export const LARGE_FILENAME = '10MB-test-file';

export const SMALL_FILEPATH = getResourceFilepath(SMALL_FILENAME);
export const LARGE_FILEPATH = getResourceFilepath(LARGE_FILENAME);

export const API_URL = process.env.SERVER_URL ?? 'http://localhost:8080';
export const API_FILE_ENDPOINT = `${API_URL}/files`;
export const API_INFO_ENDPOINT = `${API_URL}/info`;

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

export type ManifestV1 = {
  meta_destination_id: string;
  meta_ext_event: string;
  filename: string;
};

export type MetadataV1 = {
  manifest: ManifestV1;
  delivery_targets: string[];
};

export type ManifestV2 = {
  version: string;
  data_stream_id: string;
  data_stream_route: string;
  received_filename: string;
  sender_id: string;
  data_producer_id: string;
  jurisdiction: string;
};

export type MetadataV2 = {
  manifest: ManifestV2;
  delivery_targets: string[];
};

export type ManifestResponse = (ManifestV1 | ManifestV2) & {
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

export function getValidMetadataV1(): MetadataV1[] {
  const filepath = getResourceFilepath('valid_metadata_v1.json');
  return JSON.parse(readFileSync(filepath).toString());
}

export function getValidMetadataV2(): MetadataV2[] {
  const filepath = getResourceFilepath('valid_metadata_v2.json');
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
