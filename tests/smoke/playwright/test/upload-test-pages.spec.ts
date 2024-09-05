import { expect, test } from '@playwright/test';

test.describe.configure({ mode: 'parallel' });

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
    test.describe("Manifest UI endpoint", () => {
        test(`Data stream: ${dataStream} / Route: ${route} displays the appropriate UI fields`, async ({ page }) => {
            await page.goto(`/manifest?data_stream=${dataStream}&data_stream_route=${route}`);
            const title = page.locator('h1');
            await expect(title).toHaveText('Please fill in the sender manifest form for your file');
            // TODO: Add more assertions on individual manifest page elements
            const nextButton = page.getByRole('button')
            await expect(nextButton).toHaveText('Next')
        })
    })    
});
