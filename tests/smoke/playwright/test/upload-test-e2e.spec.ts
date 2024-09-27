import { expect, test } from '@playwright/test';

test.describe("Upload End to End Tests", () => {
    test(`Can upload a basic file to a default data stream`, async ({ page }) => {
        const dataStream = 'dextesting';
        const route = 'testevent1';

        await page.goto(`/`);
        await page.getByLabel('Data Stream', {exact: true}).fill(dataStream);
        await page.getByLabel('Data Stream Route').fill(route);
        await page.getByRole('button', {name: /next/i }).click();


        // ITERATE FOR EACH MANIFEST
        // ADD FAKER DATA FOR TEXT FIELDS ONLY
        await page.getByLabel('Sender Id').fill('Sender123')
        await page.getByLabel('Data Producer Id').fill('Producer123')
        await page.getByLabel('Jurisdiction').fill('Jurisdiction123')
        await page.getByLabel('Received Filename').fill('small-test-file')
        await page.getByRole('button', {name: /next/i }).click();

        const fileChooserPromise = page.waitForEvent('filechooser');
    
        await page.getByRole('button', {name: 'Browse Files'}).click(); 
        const fileChooser = await fileChooserPromise;
        await fileChooser.setFiles('../upload-files/10KB-test-file');     

        await expect(page.getByText('Upload Status: Complete')).toBeVisible();
    })
})    

