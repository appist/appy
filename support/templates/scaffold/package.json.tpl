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
    "page": "^1.11.6",
    "register-service-worker": "^1.7.1",
    "svelte": "^3.23.0",
    "svelte-i18n": "^3.0.3"
  },
  "devDependencies": {
    "@appist/webpack-preset-appy": "^0.1.1",
    "@pyoner/svelte-types": "^3.4.4-2",
    "@types/jest": "^25.2.3",
    "@types/node": "^14.0.11",
    "husky": "^4.2.5",
    "lint-staged": "^10.2.9",
    "typescript": "^3.9.5",
    "webpack-cli": "^3.3.11"
  },
  "husky": {
    "hooks": {
      "pre-commit": [
        "lint-staged"
      ]
    }
  },
  "lint-staged": {
    "**/*.{css,mjs,js,json,less,md,pug,sass,scss,svelte,ts,tsx,yml,yaml}": "npx prettier --write",
    "**/*.go": "make codecheck"
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
