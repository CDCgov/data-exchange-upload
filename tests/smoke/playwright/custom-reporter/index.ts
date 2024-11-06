import type { FullConfig, FullResult, Reporter, Suite } from '@playwright/test/reporter';
import { mkdir, writeFile } from 'fs/promises';
import { dirname } from 'path';

export interface SummaryJSONReporterOptions {
  title?: string;
  htmlReportLink: string;
  outputFilename: string;
}

export type SummaryJSONReport = {
  title: string;
  htmlReportLink: string;
  status: 'passed' | 'failed' | 'timedout' | 'interrupted';
  stats: {
    startTime: string; // Date in ISO 8601 format.,
    duration: number;
    total: number;
    expected: number;
    skipped: number;
    unexpected: number;
    flaky: number;
  };
};

class SummaryJSONReporter implements Reporter {
  private suite: Suite | undefined;

  constructor(private options: SummaryJSONReporterOptions) {
    if (typeof options.title === 'undefined') {
      this.options.title = 'Playwright Test Report';
    }
  }

  onBegin(_: FullConfig, suite: Suite) {
    this.suite = suite;
  }

  async onEnd(result: FullResult) {
    try {
      const report: SummaryJSONReport = {
        title: this.options.title ?? 'Playwright Test Report',
        htmlReportLink: this.options.htmlReportLink,
        status: result.status,
        stats: {
          startTime: result.startTime.toISOString(),
          duration: result.duration,
          total: this.suite?.allTests.length ?? 0,
          expected: 0,
          skipped: 0,
          unexpected: 0,
          flaky: 0
        }
      };
      this.suite?.allTests().forEach(test => {
        ++report.stats[test.outcome()];
      });
      const reportString = JSON.stringify(report, undefined, 2);

      await mkdir(dirname(this.options.outputFilename), { recursive: true });
      await writeFile(this.options.outputFilename, reportString);
    } catch (error) {
      console.error(error);
    }
  }
}

export default SummaryJSONReporter;
