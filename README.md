# appy

[![Build Status](https://github.com/appist/appy/workflows/Unit%20Test/badge.svg)](https://github.com/appist/appy/actions?workflow=Unit+Test)
[![GolangCI](https://golangci.com/badges/github.com/appist/appy.svg)](https://golangci.com/r/github.com/appist/appy)
[![Go Doc](http://img.shields.io/badge/godoc-reference-5272B4.svg)](http://godoc.org/github.com/appist/appy)
[![Go Report Card](https://goreportcard.com/badge/github.com/appist/appy)](https://goreportcard.com/report/github.com/appist/appy)
[![Coverage Status](https://img.shields.io/codecov/c/gh/appist/appy.svg?logo=codecov)](https://codecov.io/gh/appist/appy)

An opinionated productive web framework that helps scaling business easier.

## Features

Coming soon.

## Prerequisites

- [Go >= 1.13](https://golang.org/dl/)
- [NodeJS >= 13](https://nodejs.org/en/download/)

## Quick Start

#### Step 1: Create the project folder with go module

```sh
$ mkdir PROJECT_NAME && cd $_
$ go mod init PROJECT_NAME
```

#### Step 2: Create `main.go` with the content below

```go
package main

import (
  "github.com/appist/appy"
)

func main() {
  appy.Bootstrap()
}
```

#### Step 3: Initialize the appy's project layout

```sh
$ go run .
```

## Be A Better Engineer

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
