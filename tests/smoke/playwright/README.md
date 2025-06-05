# Upload API NodeJS TypeScript Example with Playwright Testing

Playwright is an open-source library by Microsoft for web testing and automation, enabling developers to script browser control for comprehensive web app testing. Playwright offers fast, reliable, and complex web app testing.

This project demonstrates how to use the Tus NodeJS Client for file uploads to the Upload API, along with integrating Playwright for automated end-to-end testing in a Node.js TypeScript environment.

## Setup Instructions

### Prerequisites

- Latest version of Node.js and npm installed
- Playwright requires Node.js version 12 or above

## Setup

Begin by installing the latest version of NodeJS and NPM. Then, install the Tus JS Client by running the `npm install` command in the terminal.

### Environment Variable

Set up environment variables in a file called `.env` at the root level of the repository.

```shell
SAMS_USERNAME=""
SAMS_PASSWORD=""
UPLOAD_URL=""
PS_API_URL=""
```

#### Optional Environment Variables

These are optional variables that can be set that are utilized by some tests in the suite:

| Variable           | Used In                            | Purpose
| ------------------ | ---------------------------------- | -------
| `UPLOAD_INFO_WAIT` | `upload-api-info-endpoint.spec.ts` | Sets the wait timeout for checking the upload INFO endpoint for validation

## Usage

To run the script, use `npm run build` and `npm test`. The console output should be similar to this:

```shell
5 passed
```