const htmlReportDir = process.env.HTML_REPORT_DIR ?? './test-reports/html';

export default {
  testDir: './test',
  reporter: [['html', { outputFolder: htmlReportDir, open: 'never' }]]
};
