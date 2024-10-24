import { expect, test } from '@playwright/test';

test.describe.configure({ mode: 'parallel' });

const manifests = JSON.parse(JSON.stringify(require('./manifests.json')));

test.describe('Upload Landing Page', () => {
  test('has the expected elements to start a file upload process', async ({ page }) => {
    await page.goto('/');
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
  manifests.forEach(({ dataStream, route }: { dataStream: string; route: string }) => {
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
      await expect(page.locator('body')).toContainText(`open v2/${dataStream}-${route}.json: `);
      await expect(page.locator('body')).toContainText('validation failure');
      await expect(page.locator('body')).toContainText('manifest validation config file not found');
    });
  });
});

test.describe('File Uploader Page', () => {
  test('has the expected elements to prepare to upload a file', async ({ page, baseURL }) => {
    const apiURL =
      process.env.SERVER_URL ?? baseURL?.replace('8081', '8080') ?? 'http://localhost:8080';
    const dataStream = 'dextesting';
    const route = 'testevent1';

    await page.goto(`/manifest?data_stream_id=${dataStream}&data_stream_route=${route}`);

    await page.getByLabel('Sender Id').fill('Sender123');
    await page.getByLabel('Data Producer Id').fill('Producer123');
    await page.getByLabel('Jurisdiction').fill('Jurisdiction123');
    await page.getByLabel('Received Filename').fill('small-test-file');
    await page.getByRole('button', { name: /next/i }).click();

    await expect(page.getByRole('heading', { level: 1, includeHidden: false }).nth(0)).toHaveText(
      'File Uploader'
    );
    const uploadEndpoint = page.getByRole('textbox', {
      name: 'Upload Endpoint'
    });
    // not the greatest way to interpret the endpoint value here, but this will have to work for now...
    await expect(uploadEndpoint).toHaveValue(`${apiURL}/files/`);

    const chunkSize = page.getByLabel('Chunk size (bytes)');
    const chunkSizeLabel = page.locator('label', {
      hasText: 'Chunk size (bytes)'
    });
    await expect(chunkSizeLabel).toContainText(
      'Note: Chunksize should be set on the client for uploading files of large size (1GB or over).'
    );
    await expect(chunkSize).toHaveValue('40000000');

    const parallelUploadRequests = page.getByRole('spinbutton', {
      name: 'Parallel upload requests'
    });
    await expect(parallelUploadRequests).toHaveValue('1');

    const browseFilesButton = page.getByRole('button', {
      name: 'Browse Files'
    });
    await expect(browseFilesButton).toHaveAttribute('onclick', 'files.click()');
  });
});

test.describe('Upload Status Page', () => {
  test('has the expected elements to display upload status', async ({ page, baseURL }) => {
    const apiURL =
      process.env.SERVER_URL ?? baseURL?.replace('8081', '8080') ?? 'http://localhost:8080';
    const dataStream = 'dextesting';
    const route = 'testevent1';
    const expectedFileName = 'small-test-file';
    const expectedSender = 'Sender123';
    const expectedDataProducer = 'Producer123';
    const expectedJurisdiction = 'Jurisdiction123';

    await page.goto(`/manifest?data_stream_id=${dataStream}&data_stream_route=${route}`);

    await page.getByLabel('Sender Id').fill(expectedSender);
    await page.getByLabel('Data Producer Id').fill(expectedDataProducer);
    await page.getByLabel('Jurisdiction').fill(expectedJurisdiction);
    await page.getByLabel('Received Filename').fill(expectedFileName);
    await page.getByRole('button', { name: /next/i }).click();

    const fileChooserPromise = page.waitForEvent('filechooser');
    const uploadId = page.url().split('/').slice(-1)[0];

    const uploadHeadResponsePromise = page.waitForResponse(
      response =>
        response.url() === `${apiURL}/files/${uploadId}` &&
        response.status() === 200 &&
        response.request().method() === 'HEAD'
    );

    const uploadPatchResponsePromise = page.waitForResponse(
      response =>
        response.url() === `${apiURL}/files/${uploadId}` &&
        response.status() === 204 &&
        response.request().method() === 'PATCH'
    );

    await page.getByRole('button', { name: 'Browse Files' }).click();

    const fileChooser = await fileChooserPromise;
    await fileChooser.setFiles('../upload-files/10KB-test-file');

    await expect((await uploadPatchResponsePromise).ok()).toBeTruthy();
    await expect((await uploadHeadResponsePromise).ok()).toBeTruthy();

    await page.reload();

    const fileHeaderContainer = page.locator('.file-header-container');
    await expect(fileHeaderContainer.getByRole('heading', { level: 1 }).nth(0)).toHaveText(
      expectedFileName
    );
    await expect(fileHeaderContainer.getByRole('heading', { level: 1 }).nth(1)).toHaveText(
      'Upload Status: Complete'
    );
    await expect(fileHeaderContainer).toContainText(`ID: ${uploadId}`);

    const fileDeliveriesContainer = page.locator('.file-deliveries-container');
    await expect(fileDeliveriesContainer.getByRole('heading', { level: 2 }).nth(0)).toHaveText(
      'Delivery Status'
    );

    const fileDeliveryEdavContainer = page.locator('.file-delivery-container').nth(0);
    await expect(fileDeliveryEdavContainer.getByRole('heading', { level: 2 })).toHaveText('EDAV');
    await expect(fileDeliveryEdavContainer.getByRole('heading', { level: 3 })).toHaveText(
      'Delivery Status: SUCCESS'
    );
    await expect(fileDeliveryEdavContainer).toContainText(`Location: uploads/edav/${uploadId}`);

    const fileDeliveryEhdiContainer = page.locator('.file-delivery-container').nth(1);
    await expect(fileDeliveryEhdiContainer.getByRole('heading', { level: 2 })).toHaveText('EHDI');
    await expect(fileDeliveryEhdiContainer.getByRole('heading', { level: 3 })).toHaveText(
      'Delivery Status: SUCCESS'
    );
    await expect(fileDeliveryEhdiContainer).toContainText(`Location: uploads/ehdi/${uploadId}`);

    const fileDeliveryEicrContainer = page.locator('.file-delivery-container').nth(2);
    await expect(fileDeliveryEicrContainer.getByRole('heading', { level: 2 })).toHaveText('EICR');
    await expect(fileDeliveryEicrContainer.getByRole('heading', { level: 3 })).toHaveText(
      'Delivery Status: SUCCESS'
    );
    await expect(fileDeliveryEicrContainer).toContainText(`Location: uploads/eicr/${uploadId}`);

    const fileDeliveryNcirdContainer = page.locator('.file-delivery-container').nth(3);
    await expect(fileDeliveryNcirdContainer.getByRole('heading', { level: 2 })).toHaveText('NCIRD');
    await expect(fileDeliveryNcirdContainer.getByRole('heading', { level: 3 })).toHaveText(
      'Delivery Status: SUCCESS'
    );
    await expect(fileDeliveryNcirdContainer).toContainText(`Location: uploads/ncird/${uploadId}`);

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
