import { expect, test } from '@playwright/test';
import {
  SMALL_FILENAME,
  UploadTarget,
  getFileSelection,
  getUploadTargets
} from '../resources/test-utils';

test.describe.configure({ mode: 'parallel' });

const targets: UploadTarget[] = getUploadTargets();
const fileSelection = getFileSelection(SMALL_FILENAME);

test.describe('Upload API/UI', () => {
  targets.forEach(({ dataStream, route }) => {
    test(`can use the UI to upload a file for Data stream: ${dataStream} / Route: ${route}`, async ({
      page
    }) => {
      await page.goto(`/`);
      await page.getByLabel('Data Stream', { exact: true }).fill(dataStream);
      await page.getByLabel('Data Stream Route').fill(route);
      await page.getByRole('button', { name: /next/i }).click();
      const textBoxes = await page.getByRole('textbox').all();
      for (const textbox of textBoxes) {
        await textbox.fill('Test');
      }
      await page.getByRole('button', { name: /next/i }).click();

      const fileChooserPromise = page.waitForEvent('filechooser');
      await page.getByRole('button', { name: 'Browse Files' }).click();
      const fileChooser = await fileChooserPromise;
      await fileChooser.setFiles(fileSelection);

      await expect(page.getByText('Upload Status: Complete')).toBeVisible({ timeout: 20000 });
    });
  });
});
