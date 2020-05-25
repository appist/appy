---
description: Covers how to configure the databases following 12factor principle.
---

# Configuring Databases

As we strictly follow [12factor](https://12factor.net/) principle, it is required of us to configure the databases in `./configs/.env.<ENV>` using the command below:

```bash
// Encrypt the config value using the secret key for APPY_ENV=<ENV>.
$ APPY_ENV=<ENV> go run . config:enc <KEY> <VALUE>
```

{% hint style="info" %}
Note that the above assumes that the secret key is stored in `./configs/<ENV>.key` which is used as the `APPY_MASTER_KEY` to encrypt the value.
{% endhint %}

In order to add a database called `primary`, we would configure it as shown in below:

* _**DB\_URI\_PRIMARY**_  
  Indicate the database connection URI.  
  
  **mysql** - refer to [this](https://github.com/go-sql-driver/mysql#parameters) for more parameters documentation  
  `mysql://<USERNAME>:<PASSWORD>@<HOST>:<PORT>/<SCHEMA>?multiStatements=true&parseTime=true`  
  
  **postgresql** - refer to [this](https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-PARAMKEYWORDS) for more parameters documentation  
  `postgresql://<USERNAME>:<PASSWORD>@<HOST>:<PORT>/<SCHEMA>?sslmode=false&connect_timeout=5`

  

* _**DB\_CONN\_MAX\_LIFETIME\_PRIMARY**_ Indicates the maximum amount of time a connection may be reused. By default, it is `5m`.   Note that expired connections may be closed lazily before reuse. If `d <= 0`, connections are reused forever.  
* _**DB\_MAX\_IDLE\_CONNS\_PRIMARY**_ Indicates the maximum number of connections in the idle connection pool. By default, it is `25`.   Note that _**DB\_MAX\_IDLE\_CONNS\_PRIMARY**_ will automatically be adjusted to be the same as _**DB\_MAX\_OPEN\_CONNS\_PRIMARY**_ if its value is greater than _**DB\_MAX\_IDLE\_CONNS\_PRIMARY.**_  
* _**DB\_MAX\_OPEN\_CONNS\_PRIMARY**_ Indicates the maximum number of open connections to the database. By default, it is `25`.  _****_
* _**DB\_REPLICA\_PRIMARY**_ Indicates if the database is a read replica. By default, it is `false`.  _****_
* _**DB\_SCHEMA\_MIGRATIONS\_TABLE\_PRIMARY**_ Indicates the table name for storing the schema migration versions. By default, it is `schema_migrations`.  _****_
* _**DB\_SCHEMA\_SEARCH\_PATH\_PRIMARY**_ Indicates the schema search path which is only used with `postgres` adapter. By default, it is `public`. _****_

{% hint style="info" %}
Note that the database name will be converted to camelCase. So, if you come up with `DB_URI_ANIMAL_WORLD`, you will have to access it using `app.DB("animalWorld")` or specify `masters:"animalWorld"` in the model's tag.
{% endhint %}

