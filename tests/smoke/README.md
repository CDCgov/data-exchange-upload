# Upload API Smoke Tests

These tests are used to verify the functionality of the Upload API.  They are comprised of two projects: Kotlin TestNG and Playwright.  The main reason for this is to increase test coverage across Tus upload clients.  The TestNG project uses the Java client, and the Playwright project uses the JavaScript client.  The TestNG project is the most comprehensive of the two and is currently the only one being run on our deployment pipeline.