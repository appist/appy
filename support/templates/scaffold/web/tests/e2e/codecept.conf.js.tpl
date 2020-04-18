exports.config = {
  output: "../../../tmp/e2e",
  helpers: {
    Playwright: {
      url: process.env.URL || "http://0.0.0.0:3000",
      show: false,
      browser: "firefox",
    },
  },
  include: {
    I: "./steps.js",
  },
  mocha: {},
  bootstrap: null,
  teardown: null,
  hooks: [],
  gherkin: {
    features: "./features/**/*.feature",
    steps: ["./step_definitions/welcome.js"],
  },
  plugins: {
    screenshotOnFail: {
      enabled: true,
    },
    retryFailedStep: {
      enabled: true,
    },
    stepByStepReport: {
      enabled: true,
    },
  },
  tests: "./**/*.spec.js",
  multiple: {
    default: {
      browsers: ["chromium", "firefox", "webkit"],
    },
  },
  name: "{{.projectName}}",
};
