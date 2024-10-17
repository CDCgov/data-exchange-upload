import { PlaywrightTestConfig, devices } from "@playwright/test";

const baseURL = process.env.UI_URL ?? 'http://localhost:8081';
const jsonReportFilename = process.env.TEST_REPORT_JSON ?? 'test-report.json'

const config: PlaywrightTestConfig = {
  // Specify the directory where your tests are located
  testDir: "./test",

  // Use this to change the number of browsers/contexts to run in parallel
  // Setting this to 1 will run tests serially which can help if you're seeing issues with parallel execution
  // Opt out of parallel tests on CI.
  workers: process.env.CI ? 1 : 4,

  // Fail the build on CI if you accidentally left test.only in the source code.
  forbidOnly: !!process.env.CI,

  // If a test fails, retry it additional 2 times
  // Retry on CI only.
  retries: 0,

  // Configure test timeout
  timeout: 30000,

  // Reporter to use
  reporter: process.env.CI
    ? 'github' : [
        ['list'],
        [
          'html',
          {
            outputFolder: `./test-reports/html`,
            open: 'never',
          },
        ],
        [
          'json',
          {
            outputFile: `./test-reports/${jsonReportFilename}`,
          },
        ],
      ],

  // Artifacts folder where screenshots, videos, and traces are stored.
  outputDir: './test-results',

  // Specify browser to use
  use: {
    // Specify browser to use. You can also use 'firefox' or 'webkit'.
    browserName: "chromium",

    // Specify browser launch options
    launchOptions: {
      headless: true, // Set to false if you want to see the browser UI
    },

    // Specify viewport size
    viewport: { width: 1280, height: 720 },

    // Specify the server url
    baseURL,
    
    // More options can be set here
  },

  // Add any global setup or teardown in here
  globalSetup: undefined,
  globalTeardown: undefined,

  // Configure projects for testing across multiple configurations
  projects: [
    {
      name: "Desktop Chromium",
      use: { ...devices["Desktop Chrome"] },
    },
    // More projects can be configured here
  ],
};

export default config;
