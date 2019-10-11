const chromedriver = require('chromedriver')
const geckodriver = require('geckodriver')

module.exports = {
  src_folders: ['web/tests/e2e/specs'],
  output_folder: 'web/tests/e2e/reports',
  page_objects_path: 'web/tests/e2e/page-objects',
  custom_assertions_path: 'web/tests/e2e/custom-assertions',
  custom_commands_path: 'web/tests/e2e/custom-commands',
  test_workers: false,
  test_settings: {
    default: {
      launch_url: process.env.URL || 'http://0.0.0.0:3001',
    },
    chrome: {
      desiredCapabilities: {
        browserName: 'chrome',
        chromeOptions: {
          w3c: false,
          args: ['headless'],
        },
      },
    },
    firefox: {
      desiredCapabilities: {
        browserName: 'firefox',
        alwaysMatch: {
          acceptInsecureCerts: true,
          'moz:firefoxOptions': {
            args: ['--headless'],
          },
        },
      },
      webdriver: {
        server_path: geckodriver.path,
        port: 4444,
      },
    },
  },
  webdriver: {
    start_process: true,
    port: 9515,
    server_path: chromedriver.path,
  },
}
