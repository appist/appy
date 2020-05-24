---
description: Everything you need to know to install appy and create your first application.
---

# Getting Started

## Prerequisites

* [Docker w/ Docker Compose &gt;= 19](https://www.docker.com/products/docker-desktop)
* [Go &gt;= 1.14](https://golang.org/dl/)
* [NodeJS &gt;= 13](https://nodejs.org/en/download/)
* [PostgreSQL &gt;= 11](https://www.postgresql.org/download/)
* [MySQL &gt;= 5.7](https://www.mysql.com/downloads/)

## Quick Start

### Step 1: Create the project folder with go module and git initialised.

```bash
$ mkdir <PROJECT_NAME> && cd $_ && go mod init $_ && git init
```

{% hint style="info" %}
The &lt;PROJECT\_NAME&gt; must be an alphanumeric string.
{% endhint %}

### Step 2: Create \`main.go\` with the snippet below.

```go
package main

import (
  "github.com/appist/appy"
)

func main() {
  appy.Scaffold(appy.ScaffoldOption{
    DBAdapter: "postgres", // only "mysql" and "postgres" are supported
    Description: "my first awesome app", // used in HTML's description meta tag, package.json and CLI help
  })
}
```

### Step 3: Initialize the appy's project layout.

```bash
$ go run .
```

### Step 4: Install project dependencies for backend and frontend.

```bash
$ make install
$ npm install
```

### Step 5: Setup your local environment with databases running in docker compose cluster.

```bash
$ go run . setup
```

### Step 6: Start developing your application locally.

```bash
$ go run . start
```

### Step 7: Build the application binary \(release mode\)

```text
$ go run . build
```

### Step 8: Tear down everything once you're done.

```bash
$ go run . teardown
```

{% hint style="info" %}
Now, you can execute `go run . --help` to see what `appy` built-in commands are available.
{% endhint %}

