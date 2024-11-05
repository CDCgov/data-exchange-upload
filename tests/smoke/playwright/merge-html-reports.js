var fs = require('fs');

var testFolder = process.env.TEST_REPORTS_DIR ?? './test-reports';
var stream = fs.createWriteStream(`${testFolder}/index.html`);

stream.write(`
<!DOCTYPE html>
<html>
  <head>
    <meta charset='UTF-8'>
    <meta name='color-scheme' content='dark light'>
    <meta name='viewport' content='width=device-width, initial-scale=1.0'>
    <title>Playwright Test Reports</title>
  </head>
  <body>
    <H1>Playwright Test Reports</H1>\n
`);

fs.readdir(testFolder, (err, files) => {
  files.forEach(file => {
    if (file.endsWith('.html')) {
      stream.write(`<div><a href="${file}">${file}</a></div>\n`);
    }
  });

  stream.write(`
  </body>
</html>
  `);
  stream.end();
});
