import type { FullConfig, FullResult, Reporter, Suite } from '@playwright/test/reporter';
import { createWriteStream } from 'fs';

const check = `<svg aria-hidden='true' height='16' viewBox='0 0 16 16' version='1.1' width='16' data-view-component='true' class='result-success'>
        <path fillRule='evenodd' d='M13.78 4.22a.75.75 0 010 1.06l-7.25 7.25a.75.75 0 01-1.06 0L2.22 9.28a.75.75 0 011.06-1.06L6 10.94l6.72-6.72a.75.75 0 011.06 0z'></path>
      </svg>`;

const warning = `<svg aria-hidden='true' height='16' viewBox='0 0 16 16' version='1.1' width='16' data-view-component='true' class='result-warning'>
        <path fillRule='evenodd' d='M8.22 1.754a.25.25 0 00-.44 0L1.698 13.132a.25.25 0 00.22.368h12.164a.25.25 0 00.22-.368L8.22 1.754zm-1.763-.707c.659-1.234 2.427-1.234 3.086 0l6.082 11.378A1.75 1.75 0 0114.082 15H1.918a1.75 1.75 0 01-1.543-2.575L6.457 1.047zM9 11a1 1 0 11-2 0 1 1 0 012 0zm-.25-5.25a.75.75 0 00-1.5 0v2.5a.75.75 0 001.5 0v-2.5z'></path>
      </svg>`;

const cross = `<svg class='result-failure' viewBox='0 0 16 16' version='1.1' width='16' height='16' aria-hidden='true'>
        <path fillRule='evenodd' d='M3.72 3.72a.75.75 0 011.06 0L8 6.94l3.22-3.22a.75.75 0 111.06 1.06L9.06 8l3.22 3.22a.75.75 0 11-1.06 1.06L8 9.06l-3.22 3.22a.75.75 0 01-1.06-1.06L6.94 8 3.72 4.78a.75.75 0 010-1.06z'></path>
      </svg>`;

export interface DivReporterOptions {
  title?: string;
  htmlReportLink: string;
  outputFilename: string;
}

class DivReporter implements Reporter {
  private suite: Suite | undefined;

  constructor(private options: DivReporterOptions) {
    if (typeof options.title === 'undefined') {
      this.options.title = 'Playwright Test Report';
    }
  }

  onBegin(_: FullConfig, suite: Suite) {
    this.suite = suite;
  }

  onEnd(result: FullResult) {
    try {
      const stream = createWriteStream(this.options.outputFilename);
      let icon;
      if (result.status == 'passed') {
        icon = check;
      } else if (result.status == 'failed') {
        icon = cross;
      } else {
        icon = warning;
      }

      stream.write(`
      <div class="test-report-container">
        <a href="${this.options.htmlReportLink}/index.html">
          <div class="report-header">${this.options.title}</div>
          <div class="results">
            <div class="result-item">${icon}</div>
            <div class="result-item">Start ${result.startTime}</div>
            <div class="result-item">Duration ${result.duration}</div>
          </div>
        </a>
      </div>
    `);

      stream.end();
    } catch (error) {
      console.error(error);
    }
  }
}

export default DivReporter;
