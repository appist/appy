# Contributing Guide

Hi! I'm really excited that you are interested in contributing to appy. Before submitting your contribution, please make sure to take a moment and read through the following guidelines:

- [Code of Conduct](https://github.com/appist/appy/blob/master/.github/CODE_OF_CONDUCT.md)
- [Pull Request Guidelines](#pull-request-guidelines)
- [Development Setup](#development-setup)
- [Commit Convention](#commit-convention)
- [Project Structure](#project-structure)

## Pull Request Guidelines

- The `master` branch contains the latest backward compatible changes:

  - new features: bumps the `MINOR` version
  - bug fixes: bumps `PATCH` version

- The `v[0-9]+` branch contains the next version changes which are not backward compatible which bumps the `MAJOR` version.

## Development Setup

- Install [asdf](https://asdf-vm.com/) and [Docker](https://www.docker.com/get-started).
- Run `make bootstrap` to install [Golang](https://golang.org/dl/) and [NodeJS](https://nodejs.org/en/download/releases/).
- Run `make install` to install the project dependencies.
- Run `make up` to run MySQL/PostgreSQL/ElasticSearch/Redis docker containers.
- Run `make codecheck` to run vet/lint.
- Run `make test` to run unit tests.

## Commit Convention

Commit messages should follow the [commit message convention](./COMMIT_CONVENTION.md) so that changelogs can be automatically generated.

## Project Structure

Coming soon.
