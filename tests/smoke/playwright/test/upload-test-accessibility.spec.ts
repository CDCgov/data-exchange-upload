import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

const manifests = JSON.parse(JSON.stringify(require("./manifests.json")))

test.describe.configure({ mode: 'parallel' });
const axeRuleTags = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa"];

test.describe('Upload Landing Page', () => {
    test('has accessible features when loaded', async ({ page }, testInfo) => {
        await page.goto(`/`)
        const results = await new AxeBuilder({ page })
            .withTags(axeRuleTags)
            .analyze();

        expect(results.violations).toEqual([]);
    });
});

test.describe('Upload Manifest Page', () => {
    manifests.forEach(({ dataStream, route }) => {
        test(`Checks accessibility for individual mainfest page: ${dataStream} / ${route}`, async ({ page }) => {
            await page.goto(`/manifest?data_stream=${dataStream}&data_stream_route=${route}`);
            const results = await new AxeBuilder({ page })
                .withTags(axeRuleTags)
                .analyze();
            expect(results.violations).toEqual([]);
        })
    });
});

test.describe('File Upload Page', () => {
    test(`Checks accessibliity for the upload page for the dextesting/testevent1 manifest`, async ({ page }) => {
        await page.goto(`/manifest?data_stream=dextesting&data_stream_route=testevent1`);
        await page.getByLabel('Sender Id').fill('Sender123');
        await page.getByLabel('Data Producer Id').fill('Producer123');
        await page.getByLabel('Jurisdiction').fill('Jurisdiction123');
        await page.getByLabel('Received Filename').fill('small-test-file');
        await page.getByRole('button', { name: /next/i }).click();
        await expect(page).toHaveURL(/status/)
        const results = await new AxeBuilder({ page })
            .withTags(axeRuleTags)
            .analyze();
        expect(results.violations).toEqual([]);

    })
});
