const fs = require('fs');
const path = require('path');

const testFolder = process.env.TEST_REPORTS_DIR ?? './test-reports';
const stream = fs.createWriteStream(`${testFolder}/index.html`);
const cssStyle = fs.readFileSync(path.join(__dirname, 'report.css')).toString();

stream.write(`
<!DOCTYPE html>
<html>
  <head>
    <meta charset='UTF-8'>
    <meta name='color-scheme' content='dark light'>
    <meta name='viewport' content='width=device-width, initial-scale=1.0'>
    <style>
    ${cssStyle}
    </style>
    <title>Playwright Test Reports</title>
  </head>
  <body>
    <H1>Playwright Test Reports</H1>\n
`);

fs.readdir(testFolder, (err, files) => {
  try {
    files.forEach(file => {
      const filepath = `${testFolder}/${file}`;
      const stats = fs.statSync(filepath);
      if (!stats.isDirectory() && !file.startsWith('.')) {
        stream.write(fs.readFileSync(filepath));
      }
    });
  } catch (error) {
    console.error(error);
  } finally {
    stream.write(`
    </body>
  </html>
    `);
    stream.end();
  }
});
