# appy

An opinionated productive web framework that helps scaling business easier.

###### Project Status

[![Project Status](https://img.shields.io/badge/production--ready-not%20yet-brightgreen)](https://github.com/appist/appy)

###### Build Status

[![Build Status](https://github.com/appist/appy/workflows/Code%20Check/badge.svg)](https://github.com/appist/appy/actions?workflow=Code+Check)
[![Build Status](https://github.com/appist/appy/workflows/Unit%20Test/badge.svg)](https://github.com/appist/appy/actions?workflow=Unit+Test)
[![Build Status](https://github.com/appist/appy/workflows/Examples%20-%20WeWatch/badge.svg)](https://github.com/appist/appy/actions?workflow=Examples+-+WeWatch)

###### Code Quality

[![Code Climate maintainability](https://img.shields.io/codeclimate/maintainability/appist/appy)](https://codeclimate.com/github/appist/appy/maintainability)
[![Coverage Status](https://img.shields.io/codecov/c/gh/appist/appy.svg?logo=codecov)](https://codecov.io/gh/appist/appy)
[![Go Report Card](https://goreportcard.com/badge/github.com/appist/appy)](https://goreportcard.com/report/github.com/appist/appy)

###### Documentation

[![Go Doc](http://img.shields.io/badge/godoc-reference-5272B4.svg)](http://godoc.org/github.com/appist/appy)

## Features

- Modern app development with [12factor](https://12factor.net/) methodology.
- [Highly performant](https://github.com/gin-gonic/gin/blob/master/BENCHMARKS.md) routing HTTP server using [Gin](https://github.com/gin-gonic/gin).
- [Highly performant](https://github.com/go-pg/pg/wiki/FAQ#why-go-pg) PostgreSQL ORM using [go-pg](https://github.com/go-pg/pg).
- Automatically re-compile upon file changes.
- Automatically re-generate GraphQL or gRPC boilerplate codes upon schema changes using [gqlgen](https://gqlgen.com/) or [protoc](https://github.com/protocolbuffers/protobuf).
- Server-side rendered **View** templates with [html/template](https://golang.org/pkg/html/template/).
- Client-side rendered **Progressive Web App** with [Webpack](https://webpack.js.org/) + [Svelte](https://svelte.dev/).
- Single binary support with **View** + **Progressive Web App** embedded.
- Automatically prerender client-side rendered pages for SEO using [chromedp](https://github.com/chromedp/chromedp).
- Automatically remove `Set-Cookie` response headers when the `X-API-Only: 1` request header is sent.
- Automatically set the locale based on the browser's or HTTP request's `Accept-Language` header.
- Use [Storybook](https://storybook.js.org/docs/basics/introduction/) to document/demonstrate your team's UI component style guide.

## Credits

Most of the logic details in `middleware` are from:

- https://github.com/gin-contrib
- https://github.com/gorilla

## Contribution

Please make sure to read the [Contributing Guide](https://github.com/appist/appy/blob/master/.github/CONTRIBUTING.md) before making a pull request.

Thank you to all the people who already contributed to appy!

## License

[MIT](http://opensource.org/licenses/MIT)

Copyright (c) 2019-present, Appist
