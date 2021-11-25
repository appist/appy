# Git Commit Message Convention

> This is adapted from [Angular's commit convention](https://github.com/conventional-changelog/conventional-changelog/tree/master/packages/conventional-changelog-angular).

Messages must be matched by the following regex:

```js
/^(revert: )?(feat|fix|polish|docs|style|refactor|perf|test|workflow|ci|chore|types)(\(.+\))?: .{1,50}/;
```

## Examples

Appears under "Features" header, `http` module as subheader:

```
feat(http): add `prerender` middleware
```

Appears under "Bug Fixes" header, `http` module as subheader, with a link to issue #1:

```
fix(http): fix `prerender` middleware not returning the correct HTTP response header (#1)
```

Appears under "Performance Improvements" header, `http` module as subheader:

```
perf(http): improve HTTP request concurrency handling
```

Appears under "Reverts" header:

```
revert: feat(http): add `prerender` middleware
```
