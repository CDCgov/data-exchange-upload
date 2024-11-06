import { existsSync, readdirSync, readFileSync, statSync, unlinkSync, writeFileSync } from 'fs';
import { join, resolve } from 'path';
import { SummaryJSONReport } from '../custom-reporter';

function getCssStyle(): string {
  try {
    const cssFilename = join(__dirname, 'report.css');
    if (existsSync(cssFilename)) {
      return readFileSync(cssFilename).toString() ?? '';
    }
  } catch (error: any) {
    console.error(`Could not retrieve report.css: ${error}`);
  }
  return '';
}

function getStatusIcon(status: string): string {
  if (status == 'passed') {
    return `
      <svg aria-hidden='true' height='16' viewBox='0 0 16 16' version='1.1' width='16' data-view-component='true' class='result-success'>
        <path fillRule='evenodd' d='M13.78 4.22a.75.75 0 010 1.06l-7.25 7.25a.75.75 0 01-1.06 0L2.22 9.28a.75.75 0 011.06-1.06L6 10.94l6.72-6.72a.75.75 0 011.06 0z'></path>
      </svg>`;
  }

  if (status == 'failed') {
    return `
      <svg class='result-failure' viewBox='0 0 16 16' version='1.1' width='16' height='16' aria-hidden='true'>
        <path fillRule='evenodd' d='M3.72 3.72a.75.75 0 011.06 0L8 6.94l3.22-3.22a.75.75 0 111.06 1.06L9.06 8l3.22 3.22a.75.75 0 11-1.06 1.06L8 9.06l-3.22 3.22a.75.75 0 01-1.06-1.06L6.94 8 3.72 4.78a.75.75 0 010-1.06z'></path>
      </svg>`;
  }

  return `
      <svg aria-hidden='true' height='16' viewBox='0 0 16 16' version='1.1' width='16' data-view-component='true' class='result-warning'>
        <path fillRule='evenodd' d='M8.22 1.754a.25.25 0 00-.44 0L1.698 13.132a.25.25 0 00.22.368h12.164a.25.25 0 00.22-.368L8.22 1.754zm-1.763-.707c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0114.082 15H1.918a1.75 1.75 0 01-1.543-2.575L6.457 1.047zM9 11a1 1 0 11-2 0 1 1 0 012 0zm-.25-5.25a.75.75 0 00-1.5 0v2.5a.75.75 0 001.5 0v-2.5z'></path>
      </svg>`;
}

function millisecondsToMinutesAndSeconds(msec: number): string {
  return (msec / 60000).toFixed(1) + 'm';
}

function getReportSummary(reportDir: string, file: string): string {
  try {
    const filepath = join(reportDir, file);
    if (
      filepath.endsWith('.json') &&
      !filepath.startsWith('.') &&
      existsSync(filepath) &&
      !statSync(filepath).isDirectory()
    ) {
      const results: SummaryJSONReport = JSON.parse(readFileSync(filepath).toString());
      const startTime = new Date(results.stats.startTime);
      unlinkSync(filepath);

      return `
        <div class="test-report-container" onclick="location.href='${results.htmlReportLink}/index.html';">
          <div class="report-header">${results.title}</div>
          <div class="report-body">
            ${getStatusIcon(results.status)}
            <div class="results">
              <div class="report-stats">
                <div class="stat-item">
                  All <span class="stat-count">${results.stats.total}</span>
                </div>
                <div class="stat-item">
                  Passed <span class="stat-count">${results.stats.expected}</span>
                </div>
                <div class="stat-item">
                  Failed <span class="stat-count">${results.stats.unexpected}</span>
                </div>
                <div class="stat-item">
                  Flaky <span class="stat-count">${results.stats.flaky}</span>
                </div>
                <div class="stat-item">
                  Skipped <span class="stat-count">${results.stats.skipped}</span>
                </div>
              </div>
              <div class="report-time">
                ${startTime.toLocaleString('en-US')}
                Total time: ${millisecondsToMinutesAndSeconds(results.stats.duration)}
              </div>
            </div>
          </div>
        </div>
      `;
    }
  } catch (error: any) {
    console.log(`Could not process file ${reportDir}/${file}: ${error}`);
  }
  return '';
}

function getReportSummaries(reportDir: string): string {
  try {
    const files = readdirSync(reportDir);
    if (files == null || files.length < 0) {
      return '';
    }

    return files.reduce((summaries, file) => summaries + getReportSummary(reportDir, file), '');
  } catch (error: any) {
    console.log(`Could not process summary files in ${reportDir}: ${error}`);
  }
  return '';
}

function getReportDir(): string {
  const args: string[] = process.argv.slice(2);
  let reportDir: string;

  if (args.length > 1 || (args.length == 1 && args[0] == '-h')) {
    console.log(`Usage:
      merge-html-reports <reportDir>
    `);
  }

  if (args.length == 1) {
    reportDir = args[0];
  }

  reportDir = process.env.TEST_REPORTS_DIR ?? './test-reports';
  if (reportDir.startsWith('./')) {
    reportDir = resolve(process.cwd(), reportDir);
  } else {
    reportDir = resolve(reportDir);
  }
  if (!existsSync(reportDir)) {
    console.log(`Report directory: ${reportDir} does not exist.`);
    process.exit(1);
  }
  return reportDir;
}

function createIndexFile(): void {
  try {
    const reportDir = getReportDir();
    const outputFilename: string = join(reportDir, 'index.html');
    if (existsSync(outputFilename)) {
      unlinkSync(outputFilename);
    }

    writeFileSync(
      outputFilename,
      `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style>
      ${getCssStyle()}
    </style>
    <title>Combined Playwright Test Reports</title>
  </head>
  <body>
    <main>
      <div class="report-title">Combined Playwright Test Reports</div>
      <div class="report-list">
        ${getReportSummaries(reportDir)}
      </div>
    </main>
  </body>
</html>
`
    );
    process.exit(0);
  } catch (error: any) {
    console.error(`Could not create output file: ${error}`);
    process.exit(1);
  }
}

createIndexFile();
