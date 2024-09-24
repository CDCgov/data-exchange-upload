import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

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
    [
        { dataStream: "covid", route: "all-monthly-vaccination-csv" },
        { dataStream: "covid", route: "bridge-vaccination-csv" },
        { dataStream: "dex", route: "hl7-hl7ingress" },
        { dataStream: "dextesting", route: "testevent1" },
        { dataStream: "ehdi", route: "csv" },
        { dataStream: "eicr", route: "fhir" },
        { dataStream: "generic", route: "immunization-csv" },
        { dataStream: "influenza", route: "vaccination-csv" },
        { dataStream: "ndlp", route: "covidallmonthlyvaccination" },
        { dataStream: "ndlp", route: "covidbridgevaccination" },
        { dataStream: "ndlp", route: "influenzavaccination" },
        { dataStream: "ndlp", route: "routineimmunization" },
        { dataStream: "ndlp", route: "rsvprevention" },
        { dataStream: "pulsenet", route: "localsequencefile" },
        { dataStream: "routine", route: "immunization-other" },
        { dataStream: "rsv", route: "prevention-csv" },
    ].forEach(({ dataStream, route }) => {
        test(`Checks accessibility for individual mainfest page: ${dataStream} / ${route}`, async ({ page }) => {
            await page.goto(`/`);
            await page.getByLabel('Data Stream', {exact: true}).fill(dataStream);
            await page.getByLabel('Data Stream Route').fill(route);
            await page.getByRole('button', {name: /next/i }).click();
            //await page.goto(`/manifest?data_stream=${dataStream}&data_stream_route=${route}`);
            const results = await new AxeBuilder({ page })
                .withTags(axeRuleTags)
                .analyze();
            expect(results.violations).toEqual([]);
        })
    });
});

test.describe('File Upload Page', () => {
    test(`Checks accessibliity for the upload page for the dextesting/testevent1 manifest`, async ({ page }) => {
        const dataStream = 'dextesting'
        const route = 'testevent1'
        await page.goto(`/`);
        await page.getByLabel('Data Stream', {exact: true}).fill(dataStream);
        await page.getByLabel('Data Stream Route').fill(route);
        await page.getByRole('button', {name: /next/i }).click();
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
