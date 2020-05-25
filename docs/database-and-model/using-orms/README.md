---
description: Covers how to configure models to use ORM.
---

# Using ORMs

### Naming Conventions

By default, `appy` uses some naming conventions to find out how the mapping between models and database tables should be created which will pluralise your model struct names to find the respective database table.   
  
For example, a model struct called `Book`, you should have a database table called `books`. `appy` pluralisation mechanisms are very powerful, being capable of pluralising \(and singularising\) both regular and irregular words. When using class names composed of two or more words, the model struct name should follow the Go conventions, using the **PascalCase** form, while the table name must contain the words separated by underscores. Examples:

| Model / Class | Table / Schema |
| :--- | :--- |
| `Article` | `articles` |
| `LineItem` | `line_items` |
| `Deer` | `deers` |
| `Mouse` | `mice` |
| `Person` | `people` |

### Schema Conventions

`appy` uses naming conventions for the columns in database tables, depending on the purpose of these columns:

* **Foreign Keys** These fields should be in `singularised_table_name_id` form \(e.g. `order_id`\) which is what `appy` looks for when you create associations between your models 
* **Primary Keys**  By default, `appy` will use an integer column named `id` as the table's primary key \(`bigint` for PostgreSQL and MySQL\).

There are also some optional column names that will add additional features to ORM instances:

* **created\_at** Automatically gets set to the current date and time when the record is first created. 
* **deleted\_at** Automatically gets set to the current date and time whenever the record is soft-deleted. 
* **updated\_at** Automatically gets set to the current date and time whenever the record is updated.

### Creating Models

Suppose that the `products` table was created using an SQL statement like:

```sql
CREATE TABLE products (
   id INT(11) NOT NULL auto_increment,
   name VARCHAR(255),
   PRIMARY KEY (id)
);
```

To create `Product` model, embed the `record.Model` struct and you're good to go:

```go
package model

import (
    "github.com/appist/appy/record"
    "github.com/appist/appy/support"
)

type Product struct {
    record.Model         `masters:"primary" replicas:"primaryReplica" tableName:"" timezone:"local" autoIncrement:"id" primaryKeys:"id"`
    ID   support.ZInt64  `db:"id"`
    Name support.ZString `db:"name"`
}
```

{% hint style="info" %}
Note that the `db` tag is used to identify which database column the struct field maps to.
{% endhint %}

### Using record.Model Struct Tag

In Go, it is common to use struct tag to store meta-information which can be retrieved via [reflect](https://golang.org/pkg/reflect/) package. With `appy` model, we're also utilising it to allow you to easily define the meta-information below:

| Tag | Description | Default Value | Example |
| :--- | :--- | :---: | :---: |
| masters | Indicate the master databases to be used for both read/write operations which will be selected randomly during query execution. | "" | "primary1,primary2,..." |
| replicas | Indicate the replica databases to be used for read-only operations which will be selected randomly during query execution. | "" | "replica1,replica2,..." |
| tableName | Indicate the custom table name to use instead of the inflected model struct name. \(assume that the model struct name is `Product`\) | "products" | "my\_products" |
| timezone | Indicate the timezone to use when updating `created_at`, `deleted_at` and `updated_at`. Possible values: "utc", "local". | "utc" | "local" |
| autoIncrement | Indicate which database column is auto incremented which is used to skip in the `INSERT` and `UPDATE` queries. | "" | "id" |
| primaryKeys | Indicate which database column are primary keys which is used to identify the data row in database. | "id" | "id,name" |

### Storing NULL Into Database

In Go, it can be tricky if you need to store `NULL` values into the database which you'll have to use special types. For example, if the field is type `int64`, there is no way to set it to a value which will store `NULL` into the database which is why the Go's standard library provides [SQL nullable data type](https://golang.org/pkg/database/sql/#NullInt64). However, the SQL nullable data type doesn't play well with JSON serialisation/deserialisation as shown in below:

```go
package main

import (
    "database/sql"
    "fmt"
    "time"
    
    "github.com/appist/appy/record"
    "github.com/appist/appy/support"
)

type Product struct {
    record.Model        `masters:"primary" replicas:"primaryReplica" tableName:"" timezone:"local" autoIncrement:"id" primaryKeys:"id"`
    ID   sql.NullInt64  `db:"id" json:"id"`
    Name sql.NullString `db:"name" json:"name"`
}

func main() {
    db, err := sql.Open("mysql", "<DB_URI>")
    defer db.Close()
    
    if err != nil {
        panic(err)
    }

    rows, err := db.Query("SELECT * FROM products")
    defer rows.Close()
    
    if err != nil {
        panic(err)
    }

    for rows.Next() {
        var product Product
        
        err = rows.Scan(&product.ID, &product.Name)
        
        if err != nil {
            panic(err)
        }
    
        data, _ := json.Marshal(production)
        
        // { 
        //   "id": {
        //     "Valid": true,
        //     "Value": 1,
        //   }, 
        //   "name": {
        //     "Valid": true,
        //     "Value": "Bag"
        //   }
        // }
    }
}
```

As you can see in the above, although [SQL nullable data type](https://golang.org/pkg/database/sql/#NullInt64) helps with storing `NULL` value into the database while allowing Go code to identify if it's empty value but it doesn't work well when it comes to JSON marshalling. Fortunately, `appy` does provide an easy way to work with nullable data type that works with JSON marshalling:

```go
package main

import (
    "database/sql"
    "fmt"
    "time"
    
    "github.com/appist/appy/record"
    "github.com/appist/appy/support"
)

type Product struct {
    record.Model        `masters:"primary" replicas:"primaryReplica" tableName:"" timezone:"local" autoIncrement:"id" primaryKeys:"id"`
    ID   support.ZInt64  `db:"id" json:"id"`
    Name support.ZString `db:"name" json:"name"`
}

func main() {
    db, err := sql.Open("mysql", "<DB_URI>")
    defer db.Close()
    
    if err != nil {
        panic(err)
    }

    rows, err := db.Query("SELECT * FROM products")
    defer rows.Close()
    
    if err != nil {
        panic(err)
    }

    for rows.Next() {
        var product Product
        
        err = rows.Scan(&product.ID, &product.Name)
        
        if err != nil {
            panic(err)
        }
    
        data, _ := json.Marshal(production)
        
        // { 
        //   "id": 1,
        //   "name": "Bag"
        // }
    }
}
```

### Nullable Data Types

| Type | Description | How To Initialise |
| :--- | :--- | :--- |
| support.NBool | If the value is zero, it equals to null in JSON and NULL in SQL. | support.NewNBool\(b bool\) |
| support.ZBool | If the value is zero, it equals to false in JSON and NULL in SQL. | support.NewZBool\(b bool\) |
| support.NFloat64 | If the value is zero, it equals to null in JSON and NULL in SQL. | support.NewNFloat64\(f float64\) |
| support.ZFloat64 | If the value is zero, it equals to 0 in JSON and NULL in SQL. | support.NewZFloat64\(f float64\) |
| support.NInt64 | If the value is zero, it equals to null in JSON and NULL in SQL. | support.NewNInt64\(i int64\) |
| support.ZInt64 | If the value is zero, it equals to 0 in JSON and NULL in SQL. | support.NewZInt64\(i int64\) |
| support.NString | If the value is zero, it equals to null in JSON and NULL in SQL. | support.NewNString\(s string\) |
| support.ZString | If the value is zero, it equals to "" in JSON and NULL in SQL. | support.NewZString\(s string\) |
| support.NTime | If the value is zero, it equals to null in JSON and NULL in SQL. | support.NewNTime\(t time.Time\) |
| support.ZTime | If the value is zero, it equals to "0001-01-01T00:00:00Z" in JSON and NULL in SQL. | support.NewZTime\(t time.Time\) |

