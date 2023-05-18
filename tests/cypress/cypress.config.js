const env = require("dotenv").config().parsed;
const { defineConfig } = require("cypress");

module.exports = defineConfig({
  e2e: {
    video: false,
    setupNodeEvents(on, config) {
      // implement node event listeners here
    },
  },
  env: {
    ...env,
    hideCredentials: true,
    hideCredentialsOptions: {
      body: ["username", "password"]
    }
  }
});
