import { expect, test } from '@playwright/test';
import { readFileSync } from 'fs';
import { resolve } from 'path';

test.describe.configure({ mode: 'parallel' });

const metadata = JSON.parse(JSON.stringify(require('./manifests.json')));
const filename = resolve(__dirname, '..', 'test-data', '10KB-test-file');

const fileSelected: {
  name: string;
  mimeType: string;
  buffer: Buffer;
} = {
  name: filename,
  mimeType: 'text/plain',
  buffer: readFileSync(filename)
};

test.describe('Upload API/UI', () => {
  metadata.forEach(({ dataStream, route }: { dataStream: string; route: string }) => {
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
      await fileChooser.setFiles(fileSelected);

      await expect(page.getByText('Upload Status: Complete')).toBeVisible({ timeout: 20000 });
    });
  });
});
