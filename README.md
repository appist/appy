# appy

[![Build Status](https://github.com/appist/appy/workflows/Code%20Check/badge.svg)](https://github.com/appist/appy/actions?workflow=Code+Check)
[![Build Status](https://github.com/appist/appy/workflows/Unit%20Test/badge.svg)](https://github.com/appist/appy/actions?workflow=Unit+Test)
[![Build Status](https://github.com/appist/appy/workflows/Examples%20-%20WeWatch/badge.svg)](https://github.com/appist/appy/actions?workflow=Examples+-+WeWatch)
[![Code Climate maintainability](https://img.shields.io/codeclimate/maintainability/appist/appy?style=flat-square)](https://codeclimate.com/github/appist/appy/maintainability)
[![Coverage Status](https://img.shields.io/codecov/c/gh/appist/appy.svg?logo=codecov&style=flat-square)](https://codecov.io/gh/appist/appy)
[![Go Doc](http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square)](http://godoc.org/github.com/appist/appy)
[![Go Report Card](https://goreportcard.com/badge/github.com/appist/appy?style=flat-square)](https://goreportcard.com/report/github.com/appist/appy)
![GitHub](https://img.shields.io/github/license/appist/appy.svg?style=flat-square)

An opinionated productive web framework that helps scaling business easier.

## Features

- Modern app development with [12factor](https://12factor.net/) methodology.
- Highly performant routing HTTP server using [Gin](https://github.com/gin-gonic/gin).
- Automatically re-compile upon file changes.
- Automatically re-generate GraphQL or gRPC boilerplate codes upon schema changes using [gqlgen](https://gqlgen.com/) or [protoc](https://github.com/protocolbuffers/protobuf).
- Server-side rendered **View** templates with [html/template](https://golang.org/pkg/html/template/).
- Client-side rendered **Progressive Web App** with [Webpack](https://webpack.js.org/) + [Typescript](https://www.typescriptlang.org/) + [VueJS](https://vuejs.org/).
- Automatically prerender client-side rendered pages for SEO using [chromedp](https://github.com/chromedp/chromedp).
- Automatically remove `Set-Cookie` response headers when the `X-API-Only: 1` request header is sent.
- Single binary support with **View** + **Progressive Web App** embedded.

## Credits

Most of the logic details in `middleware` are from:

- https://github.com/gin-contrib
- https://github.com/gorilla

## Contribution

Please make sure to read the [Contributing Guide](https://github.com/appist/appy/blob/master/.github/CONTRIBUTING.md) before making a pull request.

Thank you to all the people who already contributed to appy!

## License

[MIT](http://opensource.org/licenses/MIT)

Copyright (c) 2019-present, See Keat (Cayter) Goh
