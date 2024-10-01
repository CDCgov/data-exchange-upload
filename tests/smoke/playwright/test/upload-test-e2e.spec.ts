import { expect, test } from '@playwright/test';

test.describe.configure({ mode: 'parallel' });

const manifests = JSON.parse(JSON.stringify(require("./manifests.json")))

test.describe("Upload API/UI", () => {
    manifests.forEach(({ dataStream, route }) => {
        test(`can use the UI to upload a file for Data stream: ${dataStream} / Route: ${route}`, async ({ page }) => {

            await page.goto(`/`);
            await page.getByLabel('Data Stream', {exact: true}).fill(dataStream);
            await page.getByLabel('Data Stream Route').fill(route);
            await page.getByRole('button', {name: /next/i }).click();
            const textBoxes = await page.getByRole('textbox').all()
            for (const textbox of textBoxes) {
                await textbox.fill('Test')
            }
            await page.getByRole('button', {name: /next/i }).click();

            const fileChooserPromise = page.waitForEvent('filechooser');
            await page.getByRole('button', {name: 'Browse Files'}).click(); 
            const fileChooser = await fileChooserPromise;
            await fileChooser.setFiles('../upload-files/10KB-test-file');     

            await expect(page.getByText('Upload Status: Complete')).toBeVisible();
        })
    })
})    

