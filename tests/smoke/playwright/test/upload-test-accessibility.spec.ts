import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';

test.describe('Upload User Interface', () => {

    test('Checks accessiblity for the upload landing page', async ({ page }, testInfo) => {
        await page.goto(`/`)
        const results = await new AxeBuilder({ page })
            .withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa"])
            .analyze();

        expect(results.violations).toEqual([]);
    });

    [
        { dataStream: "celr", route: "csv" },
        { dataStream: "celr", route: "hl7v2" },
        { dataStream: "covid", route: "all-monthly-vaccination-csv" },
        { dataStream: "covid", route: "bridge-vaccination-csv" },
        { dataStream: "daart", route: "hl7" },
        { dataStream: "dex", route: "hl7-hl7ingress" },
        { dataStream: "dextesting", route: "testevent1" },
        { dataStream: "ehdi", route: "csv" },
        { dataStream: "eicr", route: "fhir" },
        { dataStream: "influenza", route: "vaccination-csv" },
        { dataStream: "ndlp", route: "aplhistoricaldata" },
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
            await page.goto(`/manifest?data_stream=${dataStream}&data_stream_route=${route}`);
            const results = await new AxeBuilder({ page })
                .withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa"])
                .analyze();
            expect(results.violations).toEqual([]);
        })
    });

    test(`Checks accessibliity for the upload page for the daart/hl7 manifest`, async ({ page }) => {
        await page.goto(`/manifest?data_stream=daart&data_stream_route=hl7`);
        await page.getByLabel('Sender Id').fill('Sender123');
        await page.getByLabel('Data Producer Id').fill('Producer123');
        await page.getByLabel('Jurisdiction').fill('Jurisdiction123');
        await page.getByLabel('Received Filename').fill('small-test-file');
        await page.getByLabel('Original File Timestamp').fill('Timestamp123');
        await page.getByRole('button', { name: /next/i }).click();
        await expect(page).toHaveURL(/status/)
        const results = await new AxeBuilder({ page })
            .withTags(["wcag2a", "wcag2aa", "wcag21a", "wcag21aa"])
            .analyze();
        expect(results.violations).toEqual([]);

    })
});
