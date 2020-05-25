---
description: Covers how appy stores/retrieves the configuration into/from the environment.
---

# Understanding Basics

We strictly follow [12factor](https://12factor.net/) principle by storing config in environment which allows us to run our application on different environments. What does that really mean? Let us go through the scenario below.

Imagine you have an application named `appist` connects to a PostgreSQL database which we will assume the URL looks like `postgresql://user:password@localhost:5432/appist` in the development environment and below is how you connect to the database:

```go
package main

import (
				"database/sql"
				_ "github.com/lib/pq"
)

func main() {
				db, err := sql.Open(
								"postgres", 
								"postgresql://user:password@localhost:5432/appist",
				)
				
				if err != nil {
								log.Fatal(err)
				}
				
				defer db.Close()
}
```

The above runs very well locally. However, once you're done with the development, you would very likely run the code in another environment which the database URL would be different. 

To easily manage N different database URLs for N different environments, [12factor](https://12factor.net/) suggests that we rely on the environment variable which we will update the codebase to below:

```go
package main

import (
				"database/sql"
				_ "github.com/lib/pq"
)


func main() {
				db, err := sql.Open(
								"postgres", 
								os.Getenv("DB_URL"),
				)
				
				if err != nil {
								log.Fatal(err)
				}
				
				defer db.Close()
}
```

And this is how we pass in the `DB_URL` value to the application \(for example, in `staging` environment\):

```bash
$ DB_URL=postgresql://user:password@staging.aws-rds.com:5432/appist ./appist serve
```

By following the same idea, we can extend to more config values that are being used by the application and this is basically the fundamental of how `appy` stores/retrieves the application config on different environments.

