---
description: Covers how appy manages the configuration for multiple environments securely.
---

# Managing Environments

In `appy`, all the environment variables are stored in a file and loaded into memory via `os.Setenv(key, value)`. By default, only `development` and `test` config files are created:

```bash
configs/
├── .env.development (used when APPY_ENV is "development" which is the default value)
├── .env.test        (used when APPY_ENV is "test")
├── development.key  (same as APPY_MASTER_KEY but used when APPY_ENV is "development")
└── test.key         (same as APPY_MASTER_KEY but used when APPY_ENV is "test")
```

If the codebase needs to run on other environments, such as `staging` or `production` , simply do:

```bash
// Create the config file for the specific environment.
$ touch configs/.env.<ENV>

// Generate a new secret for the specific environment.
$ echo $(go run . secret) > configs/<ENV>.key
```

{% hint style="info" %}
The `release` build only reads the secret via `APPY_MASTER_KEY` environment variable which should be passed to the built binary when running it, i.e.

**APPY\_ENV=&lt;ENV&gt; APPY\_MASTER\_KEY=&lt;SECRET\_KEY&gt; ./&lt;BINARY&gt; serve**
{% endhint %}

Now, start adding your encrypted environment variable into the corresponding environment config file. For example, adding `DB_URL` encrypted value is just as simple as doing: 

```bash
// Encrypt the config value using the secret key for APPY_ENV=<ENV>.
$ APPY_ENV=<ENV> go run . config:enc <KEY> <VALUE>

// Decrypt the config value using the secret key for APPY_ENV=<ENV>.
$ APPY_ENV=<ENV> go run . config:dec <KEY>
```

After setting up the encrypted values in the config for the new environment, simply do:

```bash
// For running the app in debug mode.
$ APPY_ENV=<ENV> go run . start

// For running the app in release mode.
$ APPY_ENV=<ENV> APPY_MASTER_KEY=<SECRET_KEY> ./<BINARY> serve
$ APPY_ENV=<ENV> APPY_MASTER_KEY=<SECRET_KEY> ./<BINARY> work
```

