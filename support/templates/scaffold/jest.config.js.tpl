module.exports = {
  collectCoverage: true,
  coverageDirectory: "tmp/coverage/web",
  globals: {
    "ts-jest": {
      diagnostics: false,
    },
  },
  transform: {
    ".+\\.(css|styl|less|sass|scss|svg|png|jpg|ttf|woff|woff2)$": "jest-transform-stub",
    "^.+\\.m?jsx?$": "babel-jest",
    "^.+\\.tsx?$": "ts-jest",
    "^.+\\.svelte$": [
      "svelte-jester",
      {
        preprocess: true,
      },
    ],
  },
  moduleFileExtensions: ["js", "jsx", "json", "svelte", "ts", "tsx", "mjs"],
  moduleNameMapper: {
    "^@assets/(.*)$": "<rootDir>/assets/$1",
    "^@/(.*)$": "<rootDir>/web/src/$1",
  },
  setupFiles: ["<rootDir>/web/tests/unit/setup.js"],
  setupFilesAfterEnv: ["@testing-library/jest-dom/extend-expect"],
};
