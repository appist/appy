---
description: Everything you need to know to install appy and create your first application.
---

# Getting Started

## Prerequisites

* [Docker w/ Docker Compose  &gt;= 19](https://www.docker.com/products/docker-desktop)
* [Go &gt;= 1.14](https://golang.org/dl/)
* [NodeJS &gt;= 13](https://nodejs.org/en/download/)
* [PostgreSQL &gt;= 12](https://www.postgresql.org/download/)

## Quick Start

{% hint style="info" %}
The project scaffolding is still being built. If you're interested in experiencing appy framework, please proceed to [https://github.com/appist/appist](https://github.com/appist/appist) to try it out first.
{% endhint %}

#### Step 1: Create the project folder with go module.

```bash
// Create project folder
$ mkdir PROJECT_NAME && cd $_

// Initialize go modules for the project
$ go mod init PROJECT_NAME
```

#### Step 2: Create \`main.go\` with the snippet below.

```go
package main

import (
  "github.com/appist/appy"
)

func main() {
  appy.Bootstrap()
}
```

#### Step 3: Initialize the appy's project layout.

```bash
// Start generating the project skeleton
$ go run .
```



