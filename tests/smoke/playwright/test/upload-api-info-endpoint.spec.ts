import { expect, test } from '@playwright/test';
import {
  validateDeliveries,
  validateFileInfo,
  validateManifest,
  validateUploadStatus
} from '../resources/test-functions';
import {
  API_FILE_ENDPOINT,
  API_INFO_ENDPOINT,
  getFileSizeBytes,
  getTestConfigV2,
  getValidMetadataV1,
  getValidMetadataV2,
  InfoResponse,
  MetadataV1,
  MetadataV2,
  SMALL_FILENAME,
  SMALL_FILEPATH
} from '../resources/test-utils';
import tusClient from '../tus-playwright';

test.describe.configure({ mode: 'parallel' });

test.describe('Info Endpoint', { tag: ['@api', '@info'] }, () => {
  const filepath = SMALL_FILEPATH;
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
        const uploadResponse = await context.upload(filepath, config?.manifest);
        uploadResponse.assertSuccess();

        const uploadId = uploadResponse.getUploadId();
        expect(uploadId).not.toBeNull();
        const uploadUrlId = uploadResponse.getUploadUrlId();
        expect(uploadUrlId).not.toBeNull();

        // wait 10 seconds for deliveries to complete
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
        const uploadResponse = await context.upload(filepath, config?.manifest);
        uploadResponse.assertSuccess();

        const uploadId = uploadResponse.getUploadId();
        expect(uploadId).not.toBeNull();
        const uploadUrlId = uploadResponse.getUploadUrlId();
        expect(uploadUrlId).not.toBeNull();

        // wait 10 seconds for deliveries to complete
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
    test('should display Initiated when the upload is initiated', async ({ page, request }) => {
      const uploader = context.newUploadContext(filepath, testConfig?.manifest, {
        shouldPauseInitialized: true
      });

      let uploadResponse = await uploader.upload();
      expect(uploadResponse.getUploadStatus()).toEqual('Initiated');

      const uploadId = uploadResponse.getUploadId();
      expect(uploadId).not.toBeNull();
      const uploadUrlId = uploadResponse.getUploadUrlId();
      expect(uploadUrlId).not.toBeNull();

      let response = await request.get(`${API_INFO_ENDPOINT}/${uploadUrlId}`);
      expect(response.ok()).toBeTruthy();

      let infoResponse: InfoResponse = await response.json();
      expect(infoResponse).not.toBeNull();
      validateUploadStatus(infoResponse.upload_status, 'Initiated');

      uploadResponse = await uploader.upload();

      // wait 10 seconds for deliveries to complete
      await page.waitForTimeout(10000);
      response = await request.get(`${API_INFO_ENDPOINT}/${uploadUrlId}`);
      expect(response.ok()).toBeTruthy();

      infoResponse = await response.json();
      expect(infoResponse).not.toBeNull();

      validateFileInfo(infoResponse.file_info, fileSize);
      validateManifest(infoResponse.manifest, uploadId);
      validateUploadStatus(infoResponse.upload_status, 'Complete');
      validateDeliveries(
        infoResponse.deliveries,
        testConfig.delivery_targets,
        `${testConfig?.manifest?.received_filename}_${uploadUrlId}`
      );
    });

    test('should display In Progress when the upload is in progress', async ({ page, request }) => {
      const uploader = context.newUploadContext(filepath, testConfig?.manifest, {
        chunkSize: 100,
        shouldPauseInProgress: true
      });

      let uploadResponse = await uploader.upload();
      expect(uploadResponse.getUploadStatus()).toEqual('In Progress');

      const uploadId = uploadResponse.getUploadId();
      expect(uploadId).not.toBeNull();
      const uploadUrlId = uploadResponse.getUploadUrlId();
      expect(uploadUrlId).not.toBeNull();

      let response = await request.get(`${API_INFO_ENDPOINT}/${uploadUrlId}`);
      expect(response.ok()).toBeTruthy();

      let infoResponse: InfoResponse = await response.json();
      expect(infoResponse).not.toBeNull();
      validateUploadStatus(infoResponse.upload_status, 'In Progress');

      uploadResponse = await uploader.upload();

      // wait 10 seconds for deliveries to complete
      await page.waitForTimeout(10000);
      response = await request.get(`${API_INFO_ENDPOINT}/${uploadUrlId}`);
      expect(response.ok()).toBeTruthy();

      infoResponse = await response.json();
      expect(infoResponse).not.toBeNull();

      validateFileInfo(infoResponse.file_info, fileSize);
      validateManifest(infoResponse.manifest, uploadId);
      validateUploadStatus(infoResponse.upload_status, 'Complete');
      validateDeliveries(
        infoResponse.deliveries,
        testConfig.delivery_targets,
        `${testConfig?.manifest?.received_filename}_${uploadUrlId}`
      );
    });
  });
});
