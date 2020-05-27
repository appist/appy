# {{.projectName}}

{{.projectDesc}}

## Prerequisites

- [Go >= 1.14](https://golang.org/dl/)
- [NodeJS >= 14](https://nodejs.org/en/download/){{if eq .dbAdapter "postgres"}}
- [PostgreSQL >= 12](https://www.postgresql.org/download/){{else if eq .dbAdapter "mysql"}}
- [MySQL >= 8](https://www.mysql.com/downloads/){{end}}

## Setup Environment

- Install [Docker](https://www.docker.com/products/docker-desktop)

- Install [Homebrew](https://brew.sh/)

- Ensure `~/.bash_profile` has the snippet below:

```sh
export PATH="$(go env GOPATH)/bin:$PATH"
```

> Please run `source ~/.bash_profile` in each terminal to ensure the script take effect.

## Quick Start

```sh
// Install backend/frontend project dependencies
$ make install && npm install

// Run dc:up/db:create/db:schema:load/db:seed to setup the datastore with seed data
$ go run . setup

// Setup the locally trusted SSL certificates (optional)
$ go run . ssl:setup

// Run the golang backend/frontend server and worker for local development
$ go run . start
```
