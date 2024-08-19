import { expect, test } from '@playwright/test';

test.describe("Upload End to End Tests", () => {
    test(`Can upload a basic file to a default data stream`, async ({ page }) => {
        const dataStream = 'dextesting';
        const route = 'testevent1';

        await page.goto(`/destination`);
        await page.getByLabel('Data Stream').fill(dataStream);
        await page.getByLabel('Data Stream Route').fill(route);
        await page.getByLabel('Submit').click();

        // More TBD
        
    })
})    

