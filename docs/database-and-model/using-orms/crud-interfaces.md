---
description: Covers how to retrieve/store data from/to the database.
---

# CRUD Interfaces

If you're used to using raw SQL to find database records, then you will generally find that there are better ways to carry out the same operations in `appy` ORM. In general, it tries to insulate you from the need to use SQL in most cases and perform queries on the database for you and is compatible with MySQL and PostgreSQL. 

Regardless of which database system you're using, the ORM method format will always be the same. To better clarify how the model CRUD interfaces work, we will be using the database table and model struct below for the rest of the page.

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

// inspect the SQL
// output: 
//    INSERT INTO 
//        users (username, created_at, deleted_at, updated_at) 
//        VALUES (:username, :created_at, :deleted_at, :updated_at);
fmt.Println(app.Model(&user).Create().SQL())

// count - the affected rows count which should be 1
// err - the error from creating the user
count, err := app.Model(&user).Create().Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err := app.Model(&user, record.ModelOption{Tx: tx}).Create().Exec()

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err := app.Model(&user).Create().Exec(record.ExecOption{Context: ctx})
```

Create Multiple Users

```go
users := []User{ 
    { Username: "john" },
    { Username: "mary" },
}

// inspect the SQL
// output: 
//    INSERT INTO 
//        users (username, created_at, deleted_at, updated_at) 
//        VALUES (:username, :created_at, :deleted_at, :updated_at);
fmt.Println(app.Model(&users).Create().SQL())

// count - the affected rows count which should be 2
// err - the error from creating the users
count, err := app.Model(&users).Create().Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err := app.Model(&users, record.ModelOption{Tx: tx}).Create().Exec()

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

// with context support, will cancel the query if it doesn't return within 3 seconds
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

Find User\(s\) With Primary Key

```go
user := User{ ID: 1 }

// count - the affected rows count which should be 1 if the user exists
// err - the error from finding the user
count, err := app.Model(&user).Find().Exec()

users := []User{ 
    { ID: 1 },
    { ID: 2 },
}

// count - the affected rows count which should be 2 if the users exist
// err - the error from finding the user
count, err := app.Model(&users).Find().Exec()
```

Find User\(s\) With `Where`, `Order`, `Limit`, `Offset`

```go
now := time.Now()

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

// only the 1st row of the returned result will be filled up
var user User
count, err := app.Model(&user).Where("username = ?", "foo").Find().Exec()

user = User{}
count, err = app.Model(&user).Where("username IN (?)", []string{"foo", "bar"}).Find().Exec()

user = User{}
count, err = app.Model(&user).Where("username IN (?) AND created_at > ?", []string{"foo", "bar"}, now.Add(time.Duration(-5) * time.Second)).Find().Exec()

user = User{}
count, err = app.Model(&user).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Find().Exec()

user = User{}
count, err = app.Model(&user).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Order("created_at ASC").Limit(1).Offset(5).Find().Exec()

user = User{}
// only user.Username and user.CreatedAt will be filled up
count, err = app.Model(&user).Select("username, created_at").Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Order("created_at ASC").Limit(1).Offset(5).Find().Exec()

user = User{}
// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = app.Model(&user, record.ModelOption{Tx: tx}).Where("username = ?", "foo").Find().Exec()

user = User{}
// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = app.Model(&user).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Find().Exec(record.ExecOption{Context: ctx})

user = User{}
// use 1 of the replicas defined in the model struct tag to execute the query
count, err = app.Model(&user).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Find().Exec(record.ExecOption{UseReplica: true})

var users []User
count, err = app.Model(&users).Where("username = ?", "foo").Find().Exec()

users = []User{}
count, err = app.Model(&users).Where("username IN (?)", []string{"foo", "bar"}).Find().Exec()

users = []User{}
count, err = app.Model(&users).Where("username IN (?) AND created_at > ?", []string{"foo", "bar"}, now.Add(time.Duration(-5) * time.Second)).Find().Exec()

users = []User{}
count, err = app.Model(&users).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Find().Exec()

users = []User{}
count, err = app.Model(&users).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Order("created_at ASC").Limit(1).Offset(5).Find().Exec()

users = []User{}
// only users[i].Username and users[i].CreatedAt will be filled up
count, err = app.Model(&users).Select("username, created_at").Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Order("created_at ASC").Limit(1).Offset(5).Find().Exec()

users = []User{}
// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = app.Model(&users, record.ModelOption{Tx: tx}).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Find().Exec()

users = []User{}
// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = app.Model(&users).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Find().Exec(record.ExecOption{Context: ctx})

users = []User{}
// use 1 of the replicas defined in the model struct tag to execute the query
count, err = app.Model(&users).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Find().Exec(record.ExecOption{UseReplica: true})
```

{% hint style="info" %}
Note that we don't support multiple `Where()` to keep things simple. In case you have multiple `Where()`, the latter one would override the prior ones.
{% endhint %}

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

Delete Single User

```go

```

Delete Multiple Users

```go

```

{% hint style="info" %}
Before updating the user in the database, each of the model instance's `DeletedAt` would be updated to the current timestamp if the `DeletedAt` struct field exists which means the model instances should perform soft delete.

In addition, it will also trigger each of the model instance's ORM callbacks in the sequence below:

* BeforeValidate
* AfterValidate
* BeforeDelete
* AfterDelete
* AfterDeleteCommit / AfterRollback \(if executed within SQL transaction\)

Note: If any of the callbacks returns error, it will return immediately without executing the remaining callbacks.
{% endhint %}

