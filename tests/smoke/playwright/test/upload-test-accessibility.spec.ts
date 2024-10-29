import AxeBuilder from '@axe-core/playwright';
import { expect, test } from '@playwright/test';
import { UploadTarget, getUploadTargets } from '../resources/test-helpers';

test.describe.configure({ mode: 'parallel' });

const targets: UploadTarget[] = getUploadTargets();
const axeRuleTags = ['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'];

test.describe('Upload Landing Page', () => {
  test('has accessible features when loaded', async ({ page }) => {
    await page.goto(`/`);
    const results = await new AxeBuilder({ page }).withTags(axeRuleTags).analyze();

    expect(results.violations).toEqual([]);
  });
});

test.describe('Upload Manifest Page', () => {
  targets.forEach(({ dataStream, route }) => {
    test(`Checks accessibility for individual metadata page: ${dataStream} / ${route}`, async ({
      page
    }) => {
      await page.goto(`/manifest?data_stream_id=${dataStream}&data_stream_route=${route}`);
      const results = await new AxeBuilder({ page }).withTags(axeRuleTags).analyze();
      expect(results.violations).toEqual([]);
    });
  });
});

test.describe('File Upload Page', () => {
  test(`Checks accessability for the upload page for the dextesting/testevent1 manifest`, async ({
    page
  }) => {
    await page.goto(`/manifest?data_stream_id=dextesting&data_stream_route=testevent1`);
    await page.getByLabel('Sender Id').fill('Sender123');
    await page.getByLabel('Data Producer Id').fill('Producer123');
    await page.getByLabel('Jurisdiction').fill('Jurisdiction123');
    await page.getByLabel('Received Filename').fill('small-test-file');
    await page.getByRole('button', { name: /next/i }).click();
    await expect(page).toHaveURL(/status/);
    const results = await new AxeBuilder({ page }).withTags(axeRuleTags).analyze();
    expect(results.violations).toEqual([]);
  });
});
