# appy

[![Build Status](https://github.com/appist/appy/workflows/Unit%20Test/badge.svg)](https://github.com/appist/appy/actions?workflow=Unit+Test)
[![Go Report Card](https://goreportcard.com/badge/github.com/appist/appy)](https://goreportcard.com/report/github.com/appist/appy)
[![Coverage Status](https://img.shields.io/codecov/c/gh/appist/appy.svg?logo=codecov)](https://codecov.io/gh/appist/appy)
[![Go Doc](http://img.shields.io/badge/godoc-reference-5272B4.svg)](http://godoc.org/github.com/appist/appy)

An opinionated productive web framework that helps scaling business easier.

## Features

- Highly performant server built on top of [gin](https://github.com/gin-gonic/gin) with the built-in middleware:
  - `CSRF` protect cookies against [Cross-Site Request Forgery](<https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)>)
  - `RequestID` generate UUID v4 for every HTTP request
  - `RequestLogger` log HTTP request details (can mask sensitive query parameter)
  - `RealIP` retrieve the real client IP
  - `ResponseHeaderFilter` remove `Set-Cookie` header for mobile clients that only use APIs
  - `HealthCheck` add a configurable health check endpoint
  - `Prerender` renders the request using Chromium for SPA's SEO
  - `Gzip` compress the response content
  - `Secure` enforce various OWASP protection
  - `I18n` attach I18n to the request context for rendering
  - `ViewEngine` attach [Jet](https://github.com/CloudyKit/jet) template engine to the request context for rendering
  - `Mailer` attach Mailer to the request context for email sending
  - `SessionMngr` manage the request session in cookie/redis
  - `Recovery` recover the request from panic
- Highly powerful CLI builder built on top of [cobra](https://github.com/spf13/cobra) with the built-in commands:
  - `build` compile the static assets into go files and build the release mode binary (only available in debug build)
  - `config:dec` encrypt a config value using the key in `configs/<APPY_ENV>.key` or `APPY_MASTER_KEY`
  - `config:enc` decrypt a config value using the key in `configs/<APPY_ENV>.key` or `APPY_MASTER_KEY`
  - `dc:down` stop and remove containers, networks, images, and volumes that are defined in `.docker/docker-compose.yml`
  - `dc:restart` restart services that are defined in `.docker/docker-compose.yml`
  - `dc:up` create and start containers that are defined in `.docker/docker-compose.yml`
  - `middleware` list all the global middleware
  - `routes` list all the server-side routes
  - `serve` run the HTTP/HTTPS web server without `webpack-dev-server`
  - `ssl:setup` generate and install the locally trusted SSL certs using `mkcert`
  - `ssl:teardown` uninstall the locally trusted SSL certs using `mkcert`
  - `start` run the HTTP/HTTPS web server with `webpack-dev-server` in development watch mode (only available in debug build)
- Support [12factor](https://12factor.net/) with each environment variable being encrypted for clearer PR review
- Support automated recompile upon changes in `assets/cmd/configs/pkg/internal` folders
- Support automated code regenerate/recompile upon [GraphQL](https://graphql.org/learn/) schema changes
- Support mailer for email sending via SMTP with preview for debugging
- Support I18n for content localization
- Support [SvelteJS](https://svelte.dev/) PWA integration with production-ready webpack configuration
- Support automated request proxy to `webpack-dev-server` for development
- Support standalone binary generating that embeds `assets/configs/locales/views`
- [WIP] Support multiple databases with ORM
- [WIP] Support worker for background job processing with admin dashboard
- [WIP] Support feature toggle with admin dashboard using Redis
- [WIP] Support authentication with database/oauth2/2FA
- [WIP] Support authorization with admin dashboard using [Casbin](https://casbin.org/)

## Prerequisites

- [Go >= 1.13](https://golang.org/dl/)
- [NodeJS >=10 <=12](https://nodejs.org/en/download/)

## Quick Start

### Step 1: Create the project folder with go module

```sh
$ mkdir PROJECT_NAME && cd $_
$ go mod init PROJECT_NAME
```

### Step 2: Create `main.go` with the content below

```go
package main

import (
  "github.com/appist/appy"
)

func main() {
  appy.Bootstrap()
}
```

### Step 3: Initialize the appy's project layout

```sh
$ go run .
```

## Best Practices

- [The Twelve-Factor App](https://12factor.net)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Uber's Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [10 Things You Probably Don't Know About Go](https://talks.golang.org/2012/10things.slide)
- [Go Map Isn't Concurrent Safe](https://golangbyexample.com/go-maps-concurrency)
- [Pointer or Value Receiver?](https://flaviocopes.com/golang-methods-receivers)

## Credits

This project is heavily inspired by:

- [rails](https://github.com/rails/rails)
- [cobra](https://github.com/spf13/cobra)
- [httprouter](https://github.com/julienschmidt/httprouter)
- [gin](https://github.com/gin-gonic/gin)
- [zap](https://github.com/uber-go/zap)
- [testify](https://github.com/stretchr/testify)

## Contribution

Please make sure to read the [Contributing Guide](https://github.com/appist/appy/blob/master/.github/CONTRIBUTING.md) before making a pull request.

Thank you to all the people who already contributed to appy!

## License

[MIT](http://opensource.org/licenses/MIT)

Copyright (c) 2019-present, Appist
