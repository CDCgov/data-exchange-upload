import { expect, test } from '@playwright/test';
import {
  API_FILE_ENDPOINT,
  API_INFO_ENDPOINT,
  DATE_TIME_REGEX,
  Delivery,
  FileInfo,
  getFileSizeBytes,
  getTestConfigV2,
  getValidMetadataV1,
  getValidMetadataV2,
  InfoResponse,
  ManifestResponse,
  MetadataV1,
  MetadataV2,
  SMALL_FILENAME,
  SMALL_FILEPATH,
  UploadStatus
} from '../resources/test-helpers';
import tusClient from '../tus-playwright';

test.describe.configure({ mode: 'parallel' });

function validateDateString(dateString: string): void {
  expect(dateString).not.toBeNull();
  expect(dateString).toMatch(DATE_TIME_REGEX);
  const milliseconds = Date.parse(dateString);
  expect(milliseconds).not.toBeNaN();
  expect(milliseconds).toBeLessThan(Date.now());
}

function validateFileInfo(file_info: FileInfo | null, fileSize: number): void {
  expect(file_info).not.toBeNull();
  if (file_info) {
    validateDateString(file_info.updated_at);
    expect(file_info.size_bytes).toEqual(fileSize);
  }
}

function validateManifest(manifest: ManifestResponse | null, uploadId: string | null): void {
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

function validateUploadStatus(
  uploadStatus: UploadStatus | null,
  expectedStatus: string | null
): void {
  expect(uploadStatus).not.toBeNull();
  if (uploadStatus) {
    expect(uploadStatus.status).toEqual(expectedStatus);
    validateDateString(uploadStatus.chunk_received_at);
  }
}

function validateDeliveries(deliveries: Delivery[] | null, targets: string[], filename: string) {
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

test.describe('Info Endpoint', { tag: ['@api', '@info'] }, () => {
  const filename = SMALL_FILEPATH;
  const fileSize = getFileSizeBytes(SMALL_FILENAME);
  const context = tusClient.newContext(API_FILE_ENDPOINT);
  const metadataV1: MetadataV1[] = getValidMetadataV1();
  const metadataV2: MetadataV2[] = getValidMetadataV2();
  const testConfig: MetadataV2 = getTestConfigV2(metadataV2);

  test.describe('Metadata Config - v1', () => {
    metadataV1.forEach(config => {
      test(`has expected response for Destination id: ${config?.manifest?.meta_destination_id} / Event: ${config?.manifest?.meta_ext_event}`, async ({
        page,
        request
      }) => {
        const uploadResponse = await context.upload(filename, config?.manifest);
        uploadResponse.assertSuccess();

        const uploadId = uploadResponse.getUploadId();
        expect(uploadId).not.toBeNull();
        const uploadUrlId = uploadResponse.getUploadUrlId();
        expect(uploadUrlId).not.toBeNull();

        await page.waitForTimeout(10000);
        const response = await request.get(`${API_INFO_ENDPOINT}/${uploadUrlId}`);
        expect(response.ok()).toBeTruthy();

        const infoResponse: InfoResponse = await response.json();
        expect(infoResponse).not.toBeNull();

        validateFileInfo(infoResponse.file_info, fileSize);
        validateManifest(infoResponse.manifest, uploadId);
        validateUploadStatus(infoResponse.upload_status, 'Complete');
        validateDeliveries(
          infoResponse.deliveries,
          config.delivery_targets,
          `${config?.manifest?.filename}_${uploadUrlId}`
        );
      });
    });
  });

  test.describe('Metadata Config - v2', () => {
    metadataV2.forEach(config => {
      test(`has expected response for Data stream: ${config?.manifest?.data_stream_id} / Route: ${config?.manifest?.data_stream_route}`, async ({
        page,
        request
      }) => {
        const uploadResponse = await context.upload(filename, config?.manifest);
        uploadResponse.assertSuccess();

        const uploadId = uploadResponse.getUploadId();
        expect(uploadId).not.toBeNull();
        const uploadUrlId = uploadResponse.getUploadUrlId();
        expect(uploadUrlId).not.toBeNull();

        await page.waitForTimeout(10000);
        const response = await request.get(`${API_INFO_ENDPOINT}/${uploadUrlId}`);
        expect(response.ok()).toBeTruthy();

        const infoResponse: InfoResponse = await response.json();
        expect(infoResponse).not.toBeNull();

        validateFileInfo(infoResponse.file_info, fileSize);
        validateManifest(infoResponse.manifest, uploadId);
        validateUploadStatus(infoResponse.upload_status, 'Complete');
        validateDeliveries(
          infoResponse.deliveries,
          config.delivery_targets,
          `${config?.manifest?.received_filename}_${uploadUrlId}`
        );
      });
    });
  });

  test.describe('Upload Status', () => {
    test('should display Initiated when the upload is initiated', async ({ request }) => {
      const uploadResponse = await context.uploadInitiated(filename, testConfig?.manifest);
      expect(uploadResponse.getUploadStatus()).toEqual('Initiated');

      const uploadId = uploadResponse.getUploadId();
      expect(uploadId).not.toBeNull();
      const uploadUrlId = uploadResponse.getUploadUrlId();
      expect(uploadUrlId).not.toBeNull();

      const response = await request.get(`${API_INFO_ENDPOINT}/${uploadUrlId}`);
      expect(response.ok()).toBeTruthy();

      const infoResponse: InfoResponse = await response.json();
      expect(infoResponse).not.toBeNull();
      validateUploadStatus(infoResponse.upload_status, 'Initiated');
    });

    test('should display In Progress when the upload is in progress', async ({ page, request }) => {
      const uploadResponse = await context.uploadInProgress(filename, testConfig?.manifest);
      expect(uploadResponse.getUploadStatus()).toEqual('In Progress');

      const uploadId = uploadResponse.getUploadId();
      expect(uploadId).not.toBeNull();
      const uploadUrlId = uploadResponse.getUploadUrlId();
      expect(uploadUrlId).not.toBeNull();

      const response = await request.get(`${API_INFO_ENDPOINT}/${uploadUrlId}`);
      expect(response.ok()).toBeTruthy();

      const initiatedInfoResponse: InfoResponse = await response.json();
      expect(initiatedInfoResponse).not.toBeNull();
      validateUploadStatus(initiatedInfoResponse.upload_status, 'In Progress');
    });
  });
});
