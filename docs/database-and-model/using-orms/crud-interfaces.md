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
	email VARCHAR UNIQUE NOT NULL,
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
    Email support.ZString    `db:"email"`
    Username support.ZString `db:"username"`
    CreatedAt support.ZTime  `db:"created_at"`
    DeletedAt support.ZTime  `db:"deleted_at"`
    UpdatedAt support.ZTime  `db:"updated_at"`
}
```

## Create Records



## Query Records



## Update Records



## Delete Records



