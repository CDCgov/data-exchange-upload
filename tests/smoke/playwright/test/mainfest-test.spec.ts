import { expect, test } from '@playwright/test';

test.describe("Manifest UI endpoint", () => {
    test("Data stream: dextesting / Route: testevent1 displays the appropriate UI fields", async ({ page }) => {
        await page.goto("http://localhost:8000/manifest?data_stream=dextesting&data_stream_route=testevent1")
        const title = page.locator('h2')
        await expect(title).toHaveText('Please fill in the sender manifest form for your file')
    })
});


// [
//     { dataStream: "dextesting", route: "testevent1"}
// ].forEach(({ dataStream, route })) => {
//     test.describe("Manifest UI endpoint", () => {
//         test(`Data stream: ${dataStream} / Route: ${route} displays the appropriate UI fields`, async ({ page }) => {
//             await page.goto(`http://localhost:8000/manifest?data_stream=${dataStream}&data_stream_route=${route}`)
//             const title = page.locator('h2')
//             await expect(title).toHaveText('Please fill in the sender manifest form for your file')
//         })
//     })    
// }

