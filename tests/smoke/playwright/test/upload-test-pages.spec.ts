import { expect, test } from '@playwright/test';

[
    { dataStream: "dextesting", route: "testevent1" },
    { dataStream: "abcs", route: "csv" },
    { dataStream: "ndlp", route: "rsvprevention"},
].forEach(({ dataStream, route }) => {
    test.describe("Manifest UI endpoint", () => {
        test(`Data stream: ${dataStream} / Route: ${route} displays the appropriate UI fields`, async ({ page }) => {
            await page.goto(`/manifest?data_stream=${dataStream}&data_stream_route=${route}`);
            const title = page.locator('h2');
            await expect(title).toHaveText('Please fill in the sender manifest form for your file');
            // TODO: Add more assertions on different common page elements
            // TODO: Add custom assertions/checkers for each individual stream/route combo in the parameters above
        })
    })    
});
