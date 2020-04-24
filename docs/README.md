---
description: Everything you need to know what appy is all about and why it was built.
---

# Introduction

[![Build Status](https://github.com/appist/appy/workflows/Unit%20Test/badge.svg)](https://github.com/appist/appy/actions?workflow=Unit+Test) [![Go Report Card](https://goreportcard.com/badge/github.com/appist/appy)](https://goreportcard.com/report/github.com/appist/appy) [![Coverage Status](https://img.shields.io/codecov/c/gh/appist/appy.svg?logo=codecov)](https://codecov.io/gh/appist/appy) [![Go Doc](http://img.shields.io/badge/godoc-reference-5272B4.svg)](https://pkg.go.dev/github.com/appist/appy?tab=doc) [![Join the chat at https://gitter.im/Appist/appy](https://badges.gitter.im/Appist/appy.svg)](https://gitter.im/Appist/appy?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) [![Platform Support](https://img.shields.io/badge/platform-macos%20%7C%20linux-blue)](https://github.com/appist/appy)

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

* Support [12factor](https://12factor.net/) app
* Support Dotenv configuration\(encrypted line-by-line\) for clearer PR review
* Support [sqlx](https://github.com/jmoiron/sqlx) with database toolings
* Support I18n on both [server-side](https://github.com/nicksnyder/go-i18n) + [client-side](https://github.com/fnando/i18n-js)
* Support [Jet template engine](https://github.com/CloudyKit/jet) for server-side rendering + browser auto-reload upon changes
* Support [SvelteJS](https://svelte.dev/) PWA static resource embedding + prerendering
* Support [NightwatchJS](https://nightwatchjs.org/) with Chrome/Firefox drivers for e2e testing
* Support [GraphQL](https://graphql.org/) development via [gqlgen](https://gqlgen.com/) + [GraphQL Playground](https://github.com/prisma-labs/graphql-playground)
* Support background job processing worker via [asynq](https://github.com/hibiken/asynq) + web UI \(coming soon\)
* Support SMTP email sending with I18n support + preview web UI
* Support custom commands building via [cobra](https://github.com/spf13/cobra)
* Support local development with auto-recompile
* Support local cluster setup using docker-compose \(even with the binary!\)
* Support local HTTPS setup via [mkcert](https://github.com/FiloSottile/mkcert)

## Caveats

* Built-in HTTP server has an [issue](https://github.com/gin-gonic/gin/issues/2016) with wildcard route

## Demo

**Debug Mode - Local Development**

![](.gitbook/assets/debug.gif)

**Release Mode - Get Ready For Deployment**

![](.gitbook/assets/release.gif)

## Acknowledgement

* [asynq](https://github.com/hibiken/asynq) - For processing background jobs
* [cobra](https://github.com/spf13/cobra) - For building CLI
* [gin](https://github.com/gin-gonic/gin) - For building web server
* [gqlgen](https://gqlgen.com/) - For building GraphQL API
* [sqlx](https://github.com/jmoiron/sqlx) - For interacting with MySQL/PostgreSQL
* [testify](https://github.com/stretchr/testify) - For writing unit tests
* [zap](https://github.com/uber-go/zap) - For blazing fast, structured and leveled logging

