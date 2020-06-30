{
  "name": "{{.projectName}}",
  "description": "{{.projectDesc}}",
  "main": "web/src/main.ts",
  "scripts": {
    "build": "NODE_ENV=production npx webpack --mode=production --progress",
    "format": "npx prettier --write '**/*.{css,mjs,js,json,less,md,pug,sass,scss,svelte,ts,tsx,yml,yaml}'",
    "start": "NODE_ENV=development npx webpack-dev-server --mode=development",
    "test": "npx jest web/src",
    "test:e2e": "npx codeceptjs run-multiple default --steps --config=web/tests/e2e/codecept.conf.js",
    "test:watch": "npm test -- --watch"
  },
  "license": "UNLICENSED",
  "dependencies": {
    "page": "^1.11.6",
    "register-service-worker": "^1.7.1",
    "svelte": "^3.23.2",
    "svelte-i18n": "^3.0.4"
  },
  "devDependencies": {
    "@appist/webpack-preset-appy": "^0.1.6",
    "@types/jest": "^26.0.3",
    "@types/node": "^14.0.14",
    "husky": "^4.2.5",
    "lint-staged": "^10.2.11",
    "typescript": "^3.9.5",
    "webpack-cli": "^3.3.12"
  },
  "babel": {
    "compact": true,
    "plugins": [
      "@babel/plugin-syntax-dynamic-import"
    ],
    "presets": [
      [
        "@babel/preset-env",
        {
          "targets": {
            "node": "current"
          }
        }
      ],
      "@babel/preset-typescript"
    ]
  },
  "husky": {
    "hooks": {
      "pre-commit": [
        "lint-staged"
      ]
    }
  },
  "jest": {
    "collectCoverage": true,
    "coverageDirectory": "tmp/coverage/web",
    "coveragePathIgnorePatterns": [
      "/node_modules/",
      ".+\\.(css|styl|less|sass|scss|svg|png|jpg|jpeg|ttf|woff|woff2)$"
    ],
    "globals": {
      "ts-jest": {
        "diagnostics": false
      }
    },
    "transform": {
      ".+\\.(css|styl|less|sass|scss|svg|png|jpg|jpeg|ttf|woff|woff2)$": "jest-transform-stub",
      "^.+\\.m?jsx?$": "babel-jest",
      "^.+\\.tsx?$": "ts-jest",
      "^.+\\.svelte$": [
        "svelte-jester",
        {
          "preprocess": true
        }
      ]
    },
    "moduleFileExtensions": [
      "js",
      "jsx",
      "json",
      "svelte",
      "ts",
      "tsx",
      "mjs"
    ],
    "moduleNameMapper": {
      "^@assets/(.*)$": "<rootDir>/assets/$1",
      "^@/(.*)$": "<rootDir>/web/src/$1"
    },
    "setupFiles": [
      "<rootDir>/web/tests/unit/setup.js"
    ],
    "setupFilesAfterEnv": [
      "@testing-library/jest-dom/extend-expect"
    ]
  },
  "lint-staged": {
    "**/*.{css,mjs,js,json,less,md,pug,sass,scss,svelte,ts,tsx,yml,yaml}": "npx prettier --write",
    "**/*.go": "make codecheck"
  },
  "prettier": {
    "arrowParens": "avoid",
    "bracketSpacing": true,
    "cursorOffset": -1,
    "endOfLine": "auto",
    "htmlWhitespaceSensitivity": "css",
    "insertPragma": false,
    "jsxBracketSameLine": false,
    "jsxSingleQuote": false,
    "printWidth": 80,
    "proseWrap": "preserve",
    "quoteProps": "as-needed",
    "requirePragma": false,
    "semi": true,
    "singleQuote": false,
    "tabWidth": 2,
    "trailingComma": "es5",
    "useTabs": false
  },
  "pwa": {
    "appleStatusBarStyle": "black-translucent",
    "background": "#ffffff",
    "developerName": "",
    "developerURL": "",
    "display": "standalone",
    "dir": "auto",
    "lang": "en-US",
    "orientation": "any",
    "scope": "/",
    "start_url": "./index.html",
    "theme_color": ""
  }
}
