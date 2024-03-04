# DeX API NodeJS TypeScript Example with Playwright Testing

Playwright is an open-source library by Microsoft for web testing and automation, enabling developers to script browser control for comprehensive web app testing.
Playwright offers fast, reliable, and complex web app testing.

This project demonstrates how to use the Tus NodeJS Client for file uploads to the DeX API, along with integrating Playwright for automated end-to-end testing in a Node.js TypeScript environment.

## Setup Instructions

### Prerequisites

=> Ensure you have the latest version of Node.js and npm installed on your machine.
=> Playwright requires Node.js version 12 or above

## Setup

Begin by installing the latest version of NodeJS and NPM. Then, install the Tus JS Client by running the `npm install` command in your terminal.

### Environment Variable

Prior to starting, you must setup envirable variables in a file called `.env` at the top level (two folders up from here) of the repository.

```bash
SAMS_USERNAME=""
SAMS_PASSWORD=""
UPLOAD_URL=""
PS_API_URL=""
```

## Usage

To run the script, invoke NodeJS passing the script name as its first argument, as in `node index.js`. You should get console output similar to this:

```
npm run build
npm test

import { expect, test } from '@playwright/test';

test.describe('DeX File Upload and Response Validation', () => {
 test.beforeAll(async ({ request }) => {
     //setup actions
    }
  });
  
  test('successful file upload returns valid response', async ({ request }) => {
    //response validation
  });
});

```
