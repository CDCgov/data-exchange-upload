# DeX API NodeJS Example

This is an example of using the Tus NodeJS Client to perform a file upload to the DeX API.

## Setup

Begin by installing the latest version of NodeJS and NPM. Then, install the Tus JS Client by running the `npm install` command in your terminal.

### Environment Variable

Prior to starting, you must setup envirable variables in a file called `.env` at the top level (two folders up from here) of the repository.

```bash
ACCOUNT_USERNAME=""
ACCOUNT_PASSWORD=""
DEX_URL=""
```

## Usage

To run the script, invoke NodeJS passing the script name as its first argument, as in `node index.js`. You should get console output similar to this:

```
0 10485760 0.00%
65536 10485760 0.63%
1245184 10485760 11.88%
7995392 10485760 76.25%
8323072 10485760 79.38%
8388608 10485760 80.00%
8585216 10485760 81.88%
8781824 10485760 83.75%
9109504 10485760 86.88%
9502720 10485760 90.63%
9961472 10485760 95.00%
10485760 10485760 100.00%
Upload finished: https://apidev.cdc.gov/upload/files/1e7778bbd0963b8a74f6055cb3489b7c
```
