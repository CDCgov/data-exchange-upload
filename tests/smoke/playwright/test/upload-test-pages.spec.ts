import { expect, test } from '@playwright/test';
import {
  API_FILE_ENDPOINT,
  SMALL_FILENAME,
  UploadTarget,
  getFileSelection,
  getUploadTargets
} from '../resources/test-utils';

test.describe.configure({ mode: 'parallel' });

const targets: UploadTarget[] = getUploadTargets();
const fileSelection = getFileSelection(SMALL_FILENAME);

test.describe('Upload Landing Page', () => {
  test('has the expected elements to start a file upload process', async ({ page }) => {
    await page.goto(`/`);
    const nav = page.getByRole('navigation');
    await expect(
      nav.getByRole('link').and(nav.getByText('Skip to main content Upload'))
    ).toBeHidden();
    await expect(nav.getByRole('link').and(nav.getByText('Upload'))).toBeVisible();
    await expect(page.getByRole('heading', { level: 1 })).toHaveText('Welcome to DEX Upload');
    await expect(page.getByRole('heading', { level: 2 })).toHaveText(
      'Start the upload process by entering a data stream and route.'
    );
    await expect(page.getByRole('textbox', { name: 'Data Stream', exact: true })).toBeVisible();
    await expect(
      page.getByRole('textbox', { name: 'Data Stream Route', exact: true })
    ).toBeVisible();
    await expect(page.getByRole('button', { name: 'Next' })).toBeVisible();
  });
});

test.describe('Upload Manifest Page', () => {
  targets.forEach(({ dataStream, route }) => {
    test(`has the expected metadata elements for Data stream: ${dataStream} / Route: ${route}`, async ({
      page
    }) => {
      await page.goto(`/manifest?data_stream_id=${dataStream}&data_stream_route=${route}`);
      const nav = page.locator('nav');
      await expect(nav.getByRole('link').and(nav.getByText('Skip to main content'))).toBeHidden();
      await expect(nav.getByRole('link').and(nav.getByText('Upload'))).toBeVisible();
      const title = page.locator('h1');
      await expect(title).toHaveText('Please fill in the sender manifest form for your file');
      // TODO: Add more assertions on individual manifest page elements
      const nextButton = page.getByRole('button');
      await expect(nextButton).toHaveText('Next');
    });
  });

  [
    { dataStream: 'invalid', route: 'invalid' }
    // Need to fix the server to enable these tests, they are currently valid when they shouldn't be
    // { dataStream: 'covid-bridge', route: 'vaccination-csv' },
    // { dataStream: 'covid', route: 'bridge-vaccination-csv' },
  ].forEach(({ dataStream, route }) => {
    test(`displays an error for an invalid manifest: Data stream: ${dataStream} / Route: ${route}`, async ({
      page
    }) => {
      const errorPagePromise = page.waitForResponse(
        `/manifest?data_stream_id=${dataStream}&data_stream_route=${route}`
      );

      await page.goto(`/`);
      await page.getByLabel('Data Stream', { exact: true }).fill(dataStream);
      await page.getByLabel('Data Stream Route').fill(route);
      await page.getByRole('button', { name: /next/i }).click();
      const errorPageResponse = await errorPagePromise;

      await expect(errorPageResponse.status()).toBe(404);
      await expect(page.locator('body')).toContainText(`open v2/${dataStream}_${route}.json: `);
      await expect(page.locator('body')).toContainText('validation failure');
      await expect(page.locator('body')).toContainText('manifest validation config file not found');
    });
  });
});

test.describe('File Uploader Page', () => {
  test('has the expected elements to prepare to upload a file', async ({ page }) => {
    const dataStream = 'dextesting';
    const route = 'testevent1';

    await page.goto(`/manifest?data_stream_id=${dataStream}&data_stream_route=${route}`);

    await page.getByLabel('Sender Id').fill('Sender123');
    await page.getByLabel('Data Producer Id').fill('Producer123');
    await page.getByLabel('Jurisdiction').fill('Jurisdiction123');
    await page.getByLabel('Received Filename').fill('small-test-file');
    await page.getByRole('button', { name: /next/i }).click();

    const chunkSize = page.getByLabel('Chunk size (bytes)');
    const chunkSizeLabel = page.locator('label', { hasText: 'Chunk size (bytes)' });
    await expect(chunkSizeLabel).toContainText(
      'Note: Chunksize should be set on the client for uploading files of large size (1GB or over).'
    );
    await expect(chunkSize).toHaveValue('40000000');

    const parallelUploadRequests = page.getByRole('spinbutton', {
      name: 'Parallel upload requests'
    });
    await expect(parallelUploadRequests).toHaveValue('1');

    await expect(page.getByRole('button', { name: 'Browse Files' })).toHaveAttribute(
      'onclick',
      'files.click()'
    );
  });
});

test.describe('Upload Status Page', () => {
  test('has the expected elements to display upload status', async ({ page, baseURL }) => {
    test.setTimeout(60000);

    const dataStream = 'dextesting';
    const route = 'testevent1';
    const expectedFileName = 'small-test-file';
    const expectedSender = 'Sender123';
    const expectedDataProducer = 'Producer123';
    const expectedJurisdiction = 'Jurisdiction123';
    const targets = ['edav'];

    await page.goto(`/manifest?data_stream_id=${dataStream}&data_stream_route=${route}`);

    await page.getByLabel('Sender Id').fill(expectedSender);
    await page.getByLabel('Data Producer Id').fill(expectedDataProducer);
    await page.getByLabel('Jurisdiction').fill(expectedJurisdiction);
    await page.getByLabel('Received Filename').fill(expectedFileName);
    await page.getByRole('button', { name: /next/i }).click();

    await expect(page.getByRole('heading', { level: 1, includeHidden: false }).nth(0)).toHaveText(
      'File Uploader'
    );

    const fileChooserPromise = page.waitForEvent('filechooser');
    const uploadId = page.url().split('/').slice(-1)[0];

    const uploadHeadResponsePromise = page.waitForResponse(
      response =>
        response.url() === `${API_FILE_ENDPOINT}/${uploadId}` &&
        response.status() === 200 &&
        response.request().method() === 'HEAD'
    );

    const uploadPatchResponsePromise = page.waitForResponse(
      response =>
        response.url() === `${API_FILE_ENDPOINT}/${uploadId}` &&
        response.status() === 204 &&
        response.request().method() === 'PATCH'
    );

    await page.getByRole('button', { name: 'Browse Files' }).click();

    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles(fileSelection);

    await expect((await uploadPatchResponsePromise).ok()).toBeTruthy();
    await expect((await uploadHeadResponsePromise).ok()).toBeTruthy();

    do {
      await page.waitForTimeout(100);
      await page.reload();
    } while ((await page.locator('.file-delivery-container')?.count()) < targets.length);

    await expect(await page.locator('.file-delivery-container').count()).toEqual(targets.length);

    const fileHeaderContainer = page.locator('.file-header-container');
    await expect(fileHeaderContainer.getByRole('heading', { level: 1 }).nth(0)).toHaveText(
      expectedFileName
    );
    await expect(fileHeaderContainer.getByRole('heading', { level: 1 }).nth(1)).toHaveText(
      'Upload Status: Complete'
    );
    // TODO handle s3 issue with ID
    // await expect(fileHeaderContainer).toContainText(`ID: ${uploadId}`)

    const fileDeliveriesContainer = page.locator('.file-deliveries-container');
    await expect(fileDeliveriesContainer.getByRole('heading', { level: 2 }).nth(0)).toHaveText(
      'Delivery Status'
    );

    targets.forEach(target => {
      (async () => {
        fileDeliveriesContainer;
        const fileDeliveryContainer = fileDeliveriesContainer
          .locator('.file-delivery-container')
          .filter({ hasText: target.toUpperCase() });
        await expect(fileDeliveryContainer.getByRole('heading', { level: 2 })).toHaveText(
          target.toUpperCase()
        );
        await expect(fileDeliveryContainer.getByRole('heading', { level: 3 })).toHaveText(
          'Delivery Status: SUCCESS'
        );
        await expect(fileDeliveryContainer).toContainText(uploadId);
        // TODO handle different destination types
        // await expect(fileDeliveryContainer).toContainText(`Location: uploads/${target}/${uploadId}`)
      })();
    });

    const uploadDetailsContainer = page.locator('.file-details-container');
    await expect(uploadDetailsContainer.getByRole('heading', { level: 2 })).toHaveText(
      'Upload Details'
    );
    await expect(uploadDetailsContainer).toContainText(`File Size: 10 KB`);
    await expect(uploadDetailsContainer).toContainText(`Sender ID: ${expectedSender}`);
    await expect(uploadDetailsContainer).toContainText(`Producer ID: ${expectedDataProducer}`);
    await expect(uploadDetailsContainer).toContainText(`Stream ID: ${dataStream}`);
    await expect(uploadDetailsContainer).toContainText(`Stream Route: ${route}`);
    await expect(uploadDetailsContainer).toContainText(`Jurisdiction: ${expectedJurisdiction}`);
  });
});
