{
  "name": "{{.projectName}}",
  "description": "{{.projectDesc}}",
  "main": "web/src/main.ts",
  "scripts": {
    "build": "NODE_ENV=production npx webpack --mode=production --progress",
    "format": "npx prettier --write '**/*.{css,mjs,js,json,less,md,pug,sass,scss,svelte,ts,tsx,yml,yaml}'",
    "start": "NODE_ENV=development npx webpack-dev-server --mode=development",
    "test": "npx jest web/src",
    "test:e2e": "npx codeceptjs run-multiple default --steps --config=web/tests/e2e/codecept.conf.js"
  },
  "license": "UNLICENSED",
  "dependencies": {
    "bootstrap": "^4.5.0",
    "page": "^1.11.6",
    "register-service-worker": "^1.7.1",
    "svelte": "^3.23.0",
    "svelte-i18n": "^3.0.3"
  },
  "devDependencies": {
    "@babel/core": "^7.10.0",
    "@babel/plugin-syntax-dynamic-import": "^7.8.3",
    "@babel/preset-env": "^7.10.0",
    "@babel/preset-typescript": "^7.9.0",
    "@intervolga/optimize-cssnano-plugin": "^1.0.6",
    "@prettier/plugin-pug": "^1.4.0",
    "@testing-library/jest-dom": "^5.8.0",
    "@testing-library/svelte": "^3.0.0",
    "@types/jest": "^25.2.3",
    "@types/node": "^14.0.5",
    "autoprefixer": "^9.8.0",
    "babel-jest": "^26.0.1",
    "babel-loader": "^8.1.0",
    "case-sensitive-paths-webpack-plugin": "^2.3.0",
    "chromedriver": "^83.0.0",
    "clean-webpack-plugin": "^3.0.0",
    "codeceptjs": "^2.6.5",
    "copy-webpack-plugin": "^6.0.1",
    "css-loader": "^3.5.3",
    "favicons-webpack-plugin": "^3.0.1",
    "file-loader": "^6.0.0",
    "friendly-errors-webpack-plugin": "^1.7.0",
    "geckodriver": "^1.18.0",
    "html-webpack-plugin": "^4.3.0",
    "husky": "^4.2.5",
    "jest": "^26.0.1",
    "jest-transform-stub": "^2.0.0",
    "lint-staged": "^10.2.6",
    "mini-css-extract-plugin": "^0.9.0",
    "node-sass": "^4.14.1",
    "playwright": "^1.0.2",
    "postcss": "^7.0.31",
    "postcss-load-config": "^2.1.0",
    "postcss-loader": "^3.0.0",
    "prettier": "^2.0.5",
    "prettier-plugin-svelte": "^1.1.0",
    "pug": "^3.0.0",
    "sass": "^1.26.5",
    "style-loader": "^1.2.1",
    "svelte-jester": "^1.0.6",
    "svelte-loader": "^2.13.6",
    "svelte-preprocess": "^3.7.4",
    "ts-jest": "^26.0.0",
    "ts-loader": "^7.0.5",
    "typescript": "^3.9.3",
    "url-loader": "^4.1.0",
    "webpack": "^4.43.0",
    "webpack-bundle-analyzer": "^3.8.0",
    "webpack-cli": "^3.3.11",
    "webpack-dev-server": "^3.11.0",
    "webpack-manifest-plugin": "^2.2.0",
    "workbox-webpack-plugin": "^5.1.3"
  },
  "babel": {
    "compact": false,
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
  "browserslist": [
    "> 1%",
    "last 2 versions"
  ],
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
    "globals": {
      "ts-jest": {
        "diagnostics": false
      }
    },
    "transform": {
      ".+\\.(css|styl|less|sass|scss|svg|png|jpg|ttf|woff|woff2)$": "jest-transform-stub",
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
    "**/*.go": "make format"
  },
  "postcss": {
    "plugins": {
      "autoprefixer": {}
    }
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
    "printWidth": 120,
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
