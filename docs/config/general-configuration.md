---
description: Covers how appy manages the configuration for multiple environments securely.
---

# Manage Environments

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
$ go run . secret --length 16

// Copy the output value and place it in "configs/<ENV>.key".
$ echo <SECRET_KEY> > configs/<ENV>.key
```

Now, start adding your encrypted environment variable into the corresponding environment config file. For example, adding `DB_URL` encrypted value is just as simple as doing: 

```bash
// Encrypt the config value using the specific environment's secret key.
// In debug mode, please ensure the secret key is in "configs/<ENV>.key".
$ APPY_ENV=<ENV> go run . config:enc <VALUE>

// Copy the encrypted value into the config file for the specific environment.
$ echo DB_URL=<ENCRYPTED_VALUE> > configs/.env.<ENV>
```

After setting up the encrypted values in the config for the new environment, simply do:

```bash
// In debug mode, please ensure the secret key is in "configs/<ENV>.key".
$ APPY_ENV=<ENV> go run . start

// In release mode, please ensure the secret key from "configs/<ENV>.key" is passed in.
$ APPY_ENV=<ENV> APPY_MASTER_KEY=<SECRET_KEY> ./<BINARY> serve
$ APPY_ENV=<ENV> APPY_MASTER_KEY=<SECRET_KEY> ./<BINARY> work
```

