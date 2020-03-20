---
name: Feature Request
about: Request for a new feature
title: "[FEATURE] <YOUR FEATURE REQUEST TITLE>"
labels: enhancement
assignees: cayter
---

**Description**

As a developer, I would like to easily rotate the secret for my config so that I don't have to re-encrypt all the config values 1 by 1.

**Expected Developer Experience**

- [ ] Run a command called `go run . config:secret:rotate <OLD_SECRET> <NEW_SECRET>`
- [ ] Re-encrypt all the config values automatically without changing the file format/comments
- [ ] Display error if the `OLD_SECRET` is invalid
