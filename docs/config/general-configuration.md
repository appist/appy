---
description: >-
  Covers the configuration and initialization features available to appy
  applications.
---

# General Configuration

In `appy`, all the environment variables are stored in a file and later on loaded into memory via `os.Setenv(key, value)`. However, this can get repeatitive and lengthy when you have lots of values to manage. Hence, we provide you an easier way to store the values securely.

