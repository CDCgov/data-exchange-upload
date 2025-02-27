import { PlaywrightTestConfig, devices } from '@playwright/test';

const baseURL = process.env.UI_URL ?? 'http://localhost:8081';
const testReportDir = process.env.TEST_REPORTS_DIR ?? './test-reports';
const jsonReportFilename = `${testReportDir}/${process.env.JSON_REPORT_FILE ?? 'test-report.json'}`;
const htmlReportLink = process.env.HTML_REPORT_DIR ?? 'html';
const htmlReportDir = `${testReportDir}/${htmlReportLink}`;
const summaryJsonReportFilename = `${testReportDir}/${process.env.SUMMARY_JSON_REPORT_FILE ?? 'summary-report.json'}`;
const testTitle = process.env.TEST_TITLE ?? 'Playwright Test Report';

const config: PlaywrightTestConfig = {
  // Specify the directory where your tests are located
  testDir: './test-auth',

  // Artifacts folder where screenshots, videos, and traces are stored.
  outputDir: './test-output',

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
    ? [
        ['github'],
        ['html', { outputFolder: htmlReportDir, open: 'never' }],
        [
          './custom-reporter/index.ts',
          {
            title: testTitle,
            htmlReportLink: `./${htmlReportLink}`,
            outputFilename: summaryJsonReportFilename
          }
        ]
      ]
    : [
        ['list', { printSteps: true }],
        ['json', { outputFile: jsonReportFilename }],
        ['html', { outputFolder: htmlReportDir, open: 'never' }]
      ],

  // Specify browser to use
  use: {
    // Specify browser to use. You can also use 'firefox' or 'webkit'.
    browserName: 'chromium',

    // Specify browser launch options
    launchOptions: {
      headless: true // Set to false if you want to see the browser UI
    },

    // Specify viewport size
    viewport: { width: 1280, height: 720 },

    // Specify the server url
    baseURL

    // More options can be set here
  },

  // Add any global setup or teardown in here
  globalSetup: undefined,
  globalTeardown: undefined,

  // Configure projects for testing across multiple configurations
  projects: [
    {
      name: 'Desktop Chromium',
      use: { ...devices['Desktop Chrome'] }
    }
    // More projects can be configured here
  ],

  webServer: {
    cwd: "./resources/mock-auth-server",
    command: 'npm install && node server.js',
    url: 'http://localhost:3000/token',
    reuseExistingServer: !process.env.CI,
    stdout: 'pipe',
    stderr: 'pipe',
  },
};

export default config;
