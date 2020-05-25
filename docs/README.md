---
description: Everything you need to know what appy is all about and why it was built.
---

# Introduction

[![Build Status](https://github.com/appist/appy/workflows/Unit%20Test/badge.svg)](https://github.com/appist/appy/actions?workflow=Unit+Test) [![Vulnerabilities Check](https://github.com/appist/appy/workflows/Vulnerabilities%20Check/badge.svg)](https://github.com/appist/appy/actions?workflow=Vulnerabilities+Check) [![Go Report Card](https://goreportcard.com/badge/github.com/appist/appy)](https://goreportcard.com/report/github.com/appist/appy) [![Coverage Status](https://img.shields.io/codecov/c/gh/appist/appy.svg?logo=codecov)](https://codecov.io/gh/appist/appy) [![Go Doc](https://img.shields.io/badge/godoc-reference-5272B4.svg)](https://pkg.go.dev/github.com/appist/appy?tab=doc) [![Platform Support](https://img.shields.io/badge/platform-macos%20%7C%20linux%20%7C%20windows-blue)](https://github.com/appist/appy)

## Motivation

Every software built has a purpose, be it for commercial or non-commercial purpose, what matters most in the end of the day is: do your engineers have the most productive framework to work with that would help to speed up the overall progress for your awesome business idea?

Common problems that slow down engineers today:

* lack of a way to store application secret correctly/securely
* lack of a way to migrate SQL database correctly/automatically
* lack of a way to ensure API documentation stays close to the codebase implementation
* lack of a way to design reliable background job processing
* lack of a way to quickly setup an affordable demo environment for their peers
* lack of a way to deliver the product to 3rd party while keeping the source code secure
* and more...

All the above are happening in startups which don't use frameworks like [Django](https://www.djangoproject.com/)/[Rails](https://rubyonrails.org/)/[Laravel](https://laravel.com/) that lead their engineers to focus on building the missing parts instead of what matters to their businesses. Even if a startup is using [Django](https://www.djangoproject.com/)/[Rails](https://rubyonrails.org/)/[Laravel](https://laravel.com/), they tend to quickly get worried about the server costs and have to spend too much time on code optimisation or re-architect their infrastructure instead of feature development.

Hence, `appy` was built to help businesses, especially startups, to focus more on feature development at its early stage by having a single monolith codebase to keep the infrastructure simple so that the business growth keeps going on. And later, when the business grows to a scale where it really needs a microservices architecture, `appy` can also help with easily migrating some of the logic out from the monolith to another service that is based off of [GRPC](https://grpc.io/).

## Features

* Support encrypted/reviewable dotenv configuration
* Support CLI building via [cobra](https://github.com/spf13/cobra)
* Support DB/ORM for interacting with multiple databases \(support context execution + replica\)
* Support DB/ORM mocks for easier unit testing
* Support powerful HTTP router [gin-gonic](https://github.com/gin-gonic/gin) + [gqlgen](https://gqlgen.com/)
* Support performant logger with [zap](https://github.com/uber-go/zap)
* Support CSR view with [SvelteJS](https://svelte.dev/) PWA + static embed + SEO pre-render
* Support SSR view with I18n + browser auto-reload
* Support e2e testing with [CodeceptJS](https://codecept.io/) + [Playwright](https://github.com/microsoft/playwright)
* Support background job worker with [asynq](https://github.com/hibiken/asynq)
* Support SMTP mailer with I18n + preview web UI
* Support local development with auto-recompile
* Support local cluster setup using docker-compose \(embeddable in the binary\)
* Support local HTTPS setup with [mkcert](https://github.com/FiloSottile/mkcert)

## Caveats

* Built-in HTTP server has an [issue](https://github.com/gin-gonic/gin/issues/2016) with wildcard route

## Demo

**Debug Mode - Local Development**

![](.gitbook/assets/debug%20%281%29.gif)

**Release Mode - Get Ready For Deployment**

![](.gitbook/assets/release%20%281%29.gif)

## Acknowledgement

* [asynq](https://github.com/hibiken/asynq) - For processing background jobs
* [cobra](https://github.com/spf13/cobra) - For building CLI
* [gin](https://github.com/gin-gonic/gin) - For building web server
* [gqlgen](https://gqlgen.com/) - For building GraphQL API
* [sqlx](https://github.com/jmoiron/sqlx) - For interacting with MySQL/PostgreSQL
* [testify](https://github.com/stretchr/testify) - For writing unit tests
* [zap](https://github.com/uber-go/zap) - For blazing fast, structured and leveled logging

