---
description: Everything you need to know what appy is all about and why it was built.
---

# Introduction

### Motivation

Every software built has a purpose, be it for commercial or non-commercial purpose, what matters most in the end of the day is: do your engineers have the most productive framework to work with that would help to speed up the overall progress for your awesome business idea?

Common problems that slow down engineers today:

* lack of a way to store application secret correctly/securely
* lack of a way to migrate SQL database correctly/automatically
* lack of a way to ensure API documentation stays close to the codebase implementation
* lack of a way to design reliable background job processing
* lack of a way to quickly setup an affordable demo environment for their peers
* lack of a way to deliver the product to 3rd party while keeping the source code secure
* and more...

All the above are happening in startups which don't use frameworks like [Django](https://www.djangoproject.com/)/[Rails](https://rubyonrails.org/)/[Laravel](https://laravel.com/) which leads their engineers to focus on building the missing parts instead of what matters to their startup businesses. Even if a startup is using [Django](https://www.djangoproject.com/)/[Rails](https://rubyonrails.org/)/[Laravel](https://laravel.com/), they tend to quickly get worried about the server costs and have to spend time too much time on code optimisation instead of feature development. And that is why \`appy\` was built to help businesses, especailly startups, to focus more on feature development at their early stage.

### Features

* Follow [12factor](https://12factor.net/) 

### Caveats

* Built-in HTTP server isn't meant/optimised for pure API server's high conccurent request handling.
* Built-in HTTP server only supports [Jet template engine](https://github.com/CloudyKit/jet) for server-side rendering.
* Built-in SQL ORM only supports [PostgreSQL](https://www.postgresql.org/) but comes with [great performance](https://github.com/go-pg/pg/wiki/FAQ#why-go-pg) in Go ecosystem.
* Built-in tooling only supports macOS and Linux due to limited Go internals support on Windows.

