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
  getValidMetadata,
  InfoResponse,
  Metadata,
  SMALL_FILENAME,
  SMALL_FILEPATH
} from '../resources/test-utils';
import tusClient from '../tus-playwright';

// set the wait time for checking the info page to the ENV VAR or 15000 by default
const UPLOAD_INFO_WAIT = process.env.UPLOAD_INFO_WAIT ? parseInt(process.env.UPLOAD_INFO_WAIT) : 15000;

test.describe.configure({ mode: 'parallel' });

test.describe('Info Endpoint', { tag: ['@api', '@info'] }, () => {
  const filepath = SMALL_FILEPATH;
  const fileSize = getFileSizeBytes(SMALL_FILENAME);
  const context = tusClient.newContext(API_FILE_ENDPOINT);
  const metadata: Metadata[] = getValidMetadata();
  const testConfig: Metadata = getTestConfigV2(metadata);

  test.describe('Metadata Config', () => {
    metadata.forEach(config => {
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
        
        await page.waitForTimeout(UPLOAD_INFO_WAIT);
        
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
          `${config?.manifest?.received_filename}`
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

      await page.waitForTimeout(UPLOAD_INFO_WAIT);
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

      await page.waitForTimeout(UPLOAD_INFO_WAIT);
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
