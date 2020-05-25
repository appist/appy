---
description: >-
  Covers all the recommended practices for how a team can work more
  effectively/efficiently.
---

# Advanced Topics

### Reviewable Configuration Change

Managing configuration for different environment can be a very tedious task which can easily lead to disaster when a careless mistake is made. Hence, it's strongly recommended that there is always another pair of eyes to help reviewing with what's being changed.

### Frequent Secret Rotation

Although `appy` framework provides a convenient way for the team to easily maintain the configuration for all environment securely on git repository, but it is still strongly recommended to frequently rotate the secrets every once in a while as a good security practice by running:

```bash
$ APPY_ENV=<ENV> go run . config:secret:rotate <OLD_SECRET> <NEW_SECRET>
```

The command above will use the `OLD_SECRET` to decrypt all values in `configs/.env.<ENV>` and re-encrypt them using `NEW_SECRET`.

