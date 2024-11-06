import { expect } from '@playwright/test';
import {
  DATE_TIME_REGEX,
  Delivery,
  FileInfo,
  ManifestResponse,
  UploadStatus
} from '../resources/test-utils';

export function validateDateString(dateString: string): void {
  expect(dateString).not.toBeNull();
  expect(dateString).toMatch(DATE_TIME_REGEX);
  const milliseconds = Date.parse(dateString);
  expect(milliseconds).not.toBeNaN();
  expect(milliseconds).toBeLessThan(Date.now());
}

export function validateFileInfo(file_info: FileInfo | null, fileSize: number): void {
  expect(file_info).not.toBeNull();
  if (file_info) {
    validateDateString(file_info.updated_at);
    expect(file_info.size_bytes).toEqual(fileSize);
  }
}

export function validateManifest(manifest: ManifestResponse | null, uploadId: string | null): void {
  expect(manifest).not.toBeNull();
  if (manifest) {
    // filter out dex_ingest_datetime to test separately
    const { dex_ingest_datetime, ...filteredManifest } = manifest;
    validateDateString(dex_ingest_datetime);
    expect(filteredManifest).toEqual({
      ...filteredManifest,
      upload_id: uploadId
    });
  }
}

export function validateUploadStatus(
  uploadStatus: UploadStatus | null,
  expectedStatus: string | null
): void {
  expect(uploadStatus).not.toBeNull();
  if (uploadStatus) {
    expect(uploadStatus.status).toEqual(expectedStatus);
    validateDateString(uploadStatus.chunk_received_at);
  }
}

export function validateDeliveries(
  deliveries: Delivery[] | null,
  targets: string[],
  filename: string
) {
  expect(deliveries).not.toBeNull();
  if (deliveries) {
    expect(deliveries.length).toEqual(targets.length);
    deliveries.forEach(({ status, name, location, delivered_at, issues }) => {
      expect(status).toEqual('SUCCESS');
      expect(targets).toContain(name);
      expect(decodeURI(location)).toContain(filename);
      validateDateString(delivered_at);
      expect(issues).toBeNull();
    });
  }
}
