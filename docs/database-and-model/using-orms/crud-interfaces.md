---
description: Covers how to retrieve/store data from/to the database.
---

# CRUD Interfaces

If you're used to using raw SQL to find database records, then you will generally find that there are better ways to carry out the same operations in `appy` ORM. In general, it tries to insulate you from the need to use SQL in most cases.  
  
`appy` ORM will perform queries on the database for you and is compatible with most database systems, including MySQL and PostgreSQL. Regardless of which database system you're using, the ORM method format will always be the same. To better clarify how the model CRUD interfaces work, we will be using the database table and model struct below for the rest of the page.

**Database Table**

```sql
CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	username VARCHAR UNIQUE NOT NULL,
	created_at TIMESTAMP,
	deleted_at TIMESTAMP,
	updated_at TIMESTAMP
);
```

**Model Struct**

```go
package model

import (
    "github.com/appist/appy/record"
    "github.com/appist/appy/support"
)

type User struct {
    record.Model             `masters:"primary" replicas:"primaryReplica" autoIncrement:"id" primaryKeys:"id"`
    ID support.ZInt64        `db:"id"`
    Username support.ZString `db:"username"`
    CreatedAt support.ZTime  `db:"created_at"`
    DeletedAt support.ZTime  `db:"deleted_at"`
    UpdatedAt support.ZTime  `db:"updated_at"`
}
```

## Create Records

Create Single User

```go
user := User{ Username: "foo" }

// count - the affected rows count which should be 1
// err - the error from creating the user
count, err := app.Model(&user).Create().Exec()

// within SQL transaction
count, err := app.Model(&user, record.ModelOption{tx}).Create().Exec()

// with context support
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

count, err := app.Model(&user).Create().Exec(record.ExecOption{Context: ctx})
```

Create Multiple Users

```go
users := []User{ 
    { Username: "john" },
    { Username: "mary" },
}

// count - the affected rows count which should be 2
// err - the error from creating the users
count, err := app.Model(&users).Create().Exec()

// within SQL transaction
count, err := app.Model(&users, record.ModelOption{tx}).Create().Exec()

// with context support
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

count, err := app.Model(&users).Create().Exec(record.ExecOption{Context: ctx})
```

{% hint style="info" %}
Before creating the user in the database, each of the model instance's `CreatedAt` would be updated to the current timestamp.

In addition, it will also trigger each of the model instance's ORM callbacks in the sequence below:

* BeforeValidate
* AfterValidate
* BeforeCreate
* AfterCreate
* AfterCreateCommit / AfterRollback \(if executed within SQL transaction\)

Note: If any of the callbacks returns error, it will return the error immediately without executing the remaining callbacks.
{% endhint %}

## Query Records



## Update Records

Update Single User

```go

```

Update Multiple Users

```go

```

{% hint style="info" %}
Before updating the user in the database, each of the model instance's `UpdatedAt` would be updated to the current timestamp.

In addition, it will also trigger each of the model instance's ORM callbacks in the sequence below:

* BeforeValidate
* AfterValidate
* BeforeUpdate
* AfterUpdate
* AfterUpdateCommit / AfterRollback \(if executed within SQL transaction\)

Note: If any of the callbacks returns error, it will return immediately without executing the remaining callbacks.
{% endhint %}

## Delete Records



