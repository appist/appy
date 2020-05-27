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

Create User\(s\)

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

Find All User\(s\)

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

var user User
count, err := app.Model(&user).All().Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = app.Model(&user, record.ModelOption{Tx: tx}).All().Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = app.Model(&user).All().Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err = app.Model(&user).All().Exec(record.ExecOption{UseReplica: true})

var users []User
count, err := app.Model(&users).All().Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = app.Model(&users, record.ModelOption{Tx: tx}).All().Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = app.Model(&users).All().Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err = app.Model(&users).All().Exec(record.ExecOption{UseReplica: true})
```

Count User\(s\) With `Select`, `Where`, `Order`, `Limit`, `Offset`

```go
now := time.Now()

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

var user User
count, err := app.Model(&user).Count().Exec()

count, err = app.Model(&user).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Order("created_at ASC").Limit(10).Offset(5).Count().Exec()

count, err = app.Model(&user).Select("DISTINCT concat(id, username)").Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Order("created_at ASC").Limit(10).Offset(5).Count().Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = app.Model(&user, record.ModelOption{Tx: tx}).Count().Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = app.Model(&user).Count().Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err = app.Model(&user).Count().Exec(record.ExecOption{UseReplica: true})

var users []User
count, err := app.Model(&users).Count().Exec()

count, err = app.Model(&users).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Order("created_at ASC").Limit(10).Offset(5).Count().Exec()

count, err = app.Model(&users).Select("DISTINCT concat(id, username)").Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).Order("created_at ASC").Limit(10).Offset(5).Count().Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = app.Model(&users, record.ModelOption{Tx: tx}).Count().Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = app.Model(&users).Count().Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err = app.Model(&users).Count().Exec(record.ExecOption{UseReplica: true})
```

Find User\(s\) With Primary Key

```go
now := time.Now()

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

user := User{ ID: 1 }

// count - the affected rows count which should be 1 if the user exists
// err - the error from finding the user
count, err := app.Model(&user).Find().Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = app.Model(&user, record.ModelOption{Tx: tx}).Find().Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = app.Model(&user).Find().Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err = app.Model(&user).Find().Exec(record.ExecOption{UseReplica: true})

users := []User{
    { ID: 1 },
    { ID: 2 },
}

// count - the affected rows count which should be 2 if the users exist
// err - the error from finding the user
count, err := app.Model(&users).Find().Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = app.Model(&users, record.ModelOption{Tx: tx}).Find().Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = app.Model(&users).Find().Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err = app.Model(&users).Find().Exec(record.ExecOption{UseReplica: true})
```

Find User\(s\) With `Select`, `Where`, `Order`, `Limit`, `Offset`

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

Find User\(s\) Custom Columns

```go
now := time.Now()

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

type Result {
    Username support.ZString
    Length   support.ZInt64
}

var result Result
count, err := s.model(&user).Select("username, LENGTH(username) AS length").Where("created_at > ?", now).Group("username").Having("username != ?", "").Scan(&result).Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = s.model(&user, record.ModelOption{Tx: tx}).Select("username, LENGTH(username) AS length").Where("created_at > ?", now).Group("username").Having("username != ?", "").Scan(&result).Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = s.model(&user).Select("username, LENGTH(username) AS length").Where("created_at > ?", now).Group("username").Having("username != ?", "").Scan(&result).Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err = s.model(&user).Select("username, LENGTH(username) AS length").Where("created_at > ?", now).Group("username").Having("username != ?", "").Scan(&result).Exec(record.ExecOption{UseReplica: true})

var results []Result
count, err = s.model(&user).Select("username, LENGTH(username) AS length").Where("created_at > ?", now).Group("username").Having("username != ?", "").Scan(&results).Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = s.model(&user, record.ModelOption{Tx: tx}).Select("username, LENGTH(username) AS length").Where("created_at > ?", now).Group("username").Having("username != ?", "").Scan(&results).Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = s.model(&user).Select("username, LENGTH(username) AS length").Where("created_at > ?", now).Group("username").Having("username != ?", "").Scan(&results).Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err = s.model(&user).Select("username, LENGTH(username) AS length").Where("created_at > ?", now).Group("username").Having("username != ?", "").Scan(&results).Exec(record.ExecOption{UseReplica: true})
```

Find User\(s\) Related Data With Join Queries

```go
type Order struct {
    record.Model             `masters:"primary" replicas:"primaryReplica" autoIncrement:"id" primaryKeys:"id"`
    ID support.ZInt64        `db:"id"`
    Username support.ZString `db:"username"`
    CreatedAt support.ZTime  `db:"created_at"`
    DeletedAt support.ZTime  `db:"deleted_at"`
    UpdatedAt support.ZTime  `db:"updated_at"`
}

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

var user User

count, err := s.model(&user).Join("INNER JOIN orders o ON o.username = username").Where("id IN (?)", []int64{1, 2}).Find().Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = s.model(&user, record.ModelOption{Tx: tx}).Join("INNER JOIN orders o ON o.username = username").Where("id IN (?)", []int64{1, 2}).Find().Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err := s.model(&user).Join("INNER JOIN orders o ON o.username = username").Where("id IN (?)", []int64{1, 2}).Find().Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err := s.model(&user).Join("INNER JOIN orders o ON o.username = username").Where("id IN (?)", []int64{1, 2}).Find().Exec(record.ExecOption{UseReplica: true})
```

## Update Records

Update User\(s\) With Primary Key

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

user := User{
    ID: 1,
    Username: "foo",
}

count, err := app.model(&user).Update().Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = app.Model(&user, record.ModelOption{Tx: tx}).Update().Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = app.Model(&user).Update().Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err = app.Model(&user).Update().Exec(record.ExecOption{UseReplica: true})

users := []User{
    {
        ID: 1,
        Username: "foo",
    },
    {
        ID: 2,
        Username: "bar",
    },
}

count, err := app.model(&users).Update().Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = app.Model(&users, record.ModelOption{Tx: tx}).Update().Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = app.Model(&users).Update().Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err = app.Model(&users).Update().Exec(record.ExecOption{UseReplica: true})
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

Update All User\(s\) With `Where`

```go
now := time.Now()

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

var user User

count, err := app.model(&user).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).UpdateAll("username = ?", "foo").Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err := app.model(&user, record.ModelOption{Tx: tx}).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).UpdateAll("username = ?", "foo").Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err := app.model(&user).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).UpdateAll("username = ?", "foo").Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err := app.model(&user).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).UpdateAll("username = ?", "foo").Exec(record.ExecOption{UseReplica: true})

count, err := app.model(&user).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).UpdateAll("username = ?", "foo").Exec()

var users []User

count, err := app.model(&users).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).UpdateAll("username = ?", "foo").Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err := app.model(&users, record.ModelOption{Tx: tx}).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).UpdateAll("username = ?", "foo").Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err := app.model(&users).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).UpdateAll("username = ?", "foo").Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err := app.model(&users).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).UpdateAll("username = ?", "foo").Exec(record.ExecOption{UseReplica: true})
```

{% hint style="info" %}
Note that `UpdateAll()` doesn't trigger any ORM callbacks.
{% endhint %}

## Delete Records

Delete User\(s\) With Primary Key

```go
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

user := User{ ID: 1 }

count, err := app.model(&user).Delete().Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = app.Model(&user, record.ModelOption{Tx: tx}).Delete().Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = app.Model(&user).Delete().Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err = app.Model(&user).Delete().Exec(record.ExecOption{UseReplica: true})

users := []User{
    { ID: 1 },
    { ID: 2 },
}

count, err := app.model(&users).Delete().Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err = app.Model(&users, record.ModelOption{Tx: tx}).Delete().Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err = app.Model(&users).Delete().Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err = app.Model(&users).Delete().Exec(record.ExecOption{UseReplica: true})
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

Delete All User\(s\) With `Where` 

```go
now := time.Now()

ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()

var user User

count, err := app.model(&user).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).DeleteAll("username = ?", "foo").Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err := app.model(&user, record.ModelOption{Tx: tx}).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).DeleteAll("username = ?", "foo").Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err := app.model(&user).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).DeleteAll("username = ?", "foo").Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err := app.model(&user).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).DeleteAll("username = ?", "foo").Exec(record.ExecOption{UseReplica: true})

count, err := app.model(&user).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).DeleteAll("username = ?", "foo").Exec()

var users []User

count, err := app.model(&users).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).DeleteAll("username = ?", "foo").Exec()

// within SQL transaction, assuming the tx was created with `app.DB('...').Begin()` or `app.Model(&products).Begin()`
count, err := app.model(&users, record.ModelOption{Tx: tx}).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).DeleteAll("username = ?", "foo").Exec()

// with context support, will cancel the query if it doesn't return within 3 seconds
count, err := app.model(&users).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).DeleteAll("username = ?", "foo").Exec(record.ExecOption{Context: ctx})

// use 1 of the replicas defined in the model struct tag to execute the query
count, err := app.model(&users).Where("created_at > ? AND created_at < ?", now.Add(time.Duration(-5) * time.Second), now).DeleteAll("username = ?", "foo").Exec(record.ExecOption{UseReplica: true})
```

{% hint style="info" %}
Note that `DeleteAll()` doesn't trigger any ORM callbacks.
{% endhint %}

