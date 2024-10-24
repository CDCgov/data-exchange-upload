import { PlaywrightTestConfig, devices } from '@playwright/test';

const config: PlaywrightTestConfig = {
  // Specify the directory where your tests are located
  testDir: './test',

  // Use this to change the number of browsers/contexts to run in parallel
  // Setting this to 1 will run tests serially which can help if you're seeing issues with parallel execution
  workers: 1,

  // Configure retries for flaky tests
  retries: 0,

  // Configure test timeout
  timeout: 30000,

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

    // More options can be set here
    baseURL: 'http://localhost:8081'
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

  // Configure reporter here. 'dot', 'list', 'junit', etc.
  reporter: [['list']]
};

export default config;
