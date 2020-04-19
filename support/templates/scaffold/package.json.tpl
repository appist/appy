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
    "bootstrap": "^4.4.1",
    "page": "^1.11.5",
    "register-service-worker": "^1.7.1",
    "svelte": "^3.20.1",
    "svelte-i18n": "^3.0.3"
  },
  "devDependencies": {
    "@babel/core": "^7.9.0",
    "@babel/plugin-syntax-dynamic-import": "^7.8.3",
    "@babel/preset-env": "^7.9.5",
    "@babel/preset-typescript": "^7.9.0",
    "@intervolga/optimize-cssnano-plugin": "^1.0.6",
    "@prettier/plugin-pug": "^1.2.0",
    "@testing-library/jest-dom": "^5.5.0",
    "@testing-library/svelte": "^3.0.0",
    "@types/jest": "^25.2.1",
    "@types/node": "^13.13.0",
    "autoprefixer": "^9.7.6",
    "babel-jest": "^25.3.0",
    "babel-loader": "^8.1.0",
    "case-sensitive-paths-webpack-plugin": "^2.3.0",
    "chromedriver": "^81.0.0",
    "clean-webpack-plugin": "^3.0.0",
    "codeceptjs": "^2.6.1",
    "copy-webpack-plugin": "^5.0.5",
    "css-loader": "^3.5.2",
    "favicons-webpack-plugin": "^3.0.1",
    "file-loader": "^6.0.0",
    "friendly-errors-webpack-plugin": "^1.7.0",
    "geckodriver": "^1.18.0",
    "html-webpack-plugin": "^4.2.0",
    "husky": "^4.2.5",
    "jest": "^25.3.0",
    "jest-transform-stub": "^2.0.0",
    "lint-staged": "^10.1.5",
    "mini-css-extract-plugin": "^0.9.0",
    "node-sass": "^4.13.1",
    "playwright": "^0.13.0",
    "postcss": "^7.0.27",
    "postcss-load-config": "^2.1.0",
    "postcss-loader": "^3.0.0",
    "prettier": "^2.0.4",
    "prettier-plugin-svelte": "^0.7.0",
    "pug": "^2.0.4",
    "sass": "^1.26.3",
    "style-loader": "^1.1.4",
    "svelte-jester": "^1.0.5",
    "svelte-loader": "^2.13.6",
    "svelte-preprocess": "^3.7.1",
    "ts-jest": "^25.4.0",
    "ts-loader": "^7.0.0",
    "typescript": "^3.8.3",
    "url-loader": "^4.1.0",
    "webpack": "^4.42.1",
    "webpack-bundle-analyzer": "^3.7.0",
    "webpack-cli": "^3.3.11",
    "webpack-dev-server": "^3.10.3",
    "webpack-manifest-plugin": "^2.2.0",
    "workbox-webpack-plugin": "^5.1.2"
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
