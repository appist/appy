---
description: Covers how to evolve your database schema over time.
---

# Migrations

### Overview

Migrations are a convenient way to alter your database schema over time in a consistent way. You can think of each migration as being a new `version` of the database. A schema starts off with nothing in it, and each migration modifies it to add or remove tables, columns, or entries.  
  
`appy` knows how to update your schema along this timeline, bringing it from whatever point it is in the history to the latest version. In `debug` mode, `appy` will also update your `db/migrate/<DATABASE>/schema.go` file to match the up-to-date structure of your database.  
  
Here's an example of a migration for the database called `primary` in `db/migrate/primary/20200201165238_create_users`:

```go
package primary

import (
	"github.com/<PROJECT_NAME>/appy/record"

	"<PROJECT_NAME>/pkg/app"
)

func init() {
	db := app.DB("primary")

	if db != nil {
		err := db.RegisterMigrationTx(
			// Up migration
			func(db record.Txer) error {
				_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	email VARCHAR UNIQUE NOT NULL,
	username VARCHAR UNIQUE NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP
);`)
				return err
			},

			// Down migration
			func(db record.Txer) error {
				_, err := db.Exec(`DROP TABLE IF EXISTS users`)
				return err
			},
		)

		if err != nil {
			app.Logger.Fatal(err)
		}
	}
}

```

This migration adds a table called `users` with:

* `id` as the auto-increment primary key
* `email` and `username` as non-nullable string columns with unique constraint
* `created_at`, `deleted_at` and `updated_at` as the timestamp columns

{% hint style="info" %}
It also comes with a `down` block which would allow us to rollback the migration in case anything goes wrong.
{% endhint %}

By default, all the generated migrations are executed within a transaction. In the case where you would want to run the migration outside of a transaction, simply generate the migration file for the `primary` database by running:

```bash
$ go run . gen:migration CreateUsers --database=primary --tx=false
```

### Creating a Migration

Migrations are stored as files in the `db/migrate/<DATABASE>` directory. The name of the file is of the form `YYYYMMDDHHMMSS_create_users.go`, that is to say a UTC timestamp identifying the migration followed by an underscore followed by the name of the migration. The name of the migration should match the latter part of the file name.   
  
For example, `CreateUsers` should generate `20080906120000_create_users.go` and `AddDetailsToUsers` should generate `20080906120001_add_details_to_users.go`.   
  
`appy` uses this timestamp to determine which migration should be run and in what order, so if you're copying a migration from another application or generate a file yourself, be aware of its position in the order.  
  
Of course, calculating timestamps is no fun, so `appy` provides a generator to handle making it for you:

```bash
// Available arguments:
//
// database - indicate which database to generate the migration file for, default is primary
// tx - indicate if the migration should be executed within a transaction, default is true

$ go run . gen:migration CreateUsers
```

This will create an appropriately named empty migration:

```go
package <DATABASE>

import (
	"github.com/appist/appy/record"
	"<PROJECT_NAME>/pkg/app"
)

func init() {
	db := app.DB("<DATABASE>")

	if db != nil {
		err := db.RegisterMigrationTx(
			// Up migration
			func(db record.Txer) error {
				_, err := db.Exec(``)
				return err
			},
			// Down migration
			func(db record.Txer) error {
				_, err := db.Exec(``)
				return err
			},
		)

		if err != nil {
			app.Logger.Fatal(err)
		}
	}
}

```

From here, you can start writing writing SQL queries for the down/up migrations!

### Running Migrations

`appy` provides the `db:migrate` command to run migrations:

```bash
// Available arguments:
//
// database - indicate which database to run the migration for, default is all databases
$ go run . db:migrate --database=<DATABASE>
```

{% hint style="info" %}
Note that running the `db:migrate` command also invokes the `db:schema:dump` command, which will update your `db/<DATABASE>/schema.go` file to match the structure of your database.
{% endhint %}

### Rolling Back Migration

`appy` provides the `db:rollback` command to revert migration:

```bash
// Available arguments:
//
// database - indicate which database to rollback the migration for, default is primary
$ go run . db:rollback --database=<DATABASE>
```

{% hint style="info" %}
Note that running the `db:rollback` command also invokes the `db:schema:dump` command, which will update your `db/<DATABASE>/schema.go` file to match the structure of your database.
{% endhint %}

### List All Migrations Status

`appy` provides the `db:migrate:status` command to list all the database migration status:

```bash
// Available arguments:
//
// database - indicate which database to run the migration for, default is all databases
$ go run . db:migrate:status --database=<DATABASE>

2020-05-25T17:41:21.361+0800    INFO    [DB] SELECT COUNT(*) FROM pg_tables WHERE schemaname = public AND tablename = schema_migrations; (39.818539ms)
2020-05-25T17:41:21.370+0800    INFO    [DB] SELECT version FROM public.schema_migrations ORDER BY version ASC; (8.984293ms)

database: primary

-----------  -------------------  ---------------------------------------------------------
 Status       Migration ID         Migration File                                          
-----------  -------------------  ---------------------------------------------------------
 up           20200201165238       db/migrate/primary/20200201165238_create_users.go       

 down         20200525172040       db/migrate/primary/20200525172040_create_products.go    
-----------  -------------------  ---------------------------------------------------------
```

