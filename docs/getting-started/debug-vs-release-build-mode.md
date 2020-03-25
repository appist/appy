---
description: >-
  Covers what the debug and release mode is and how they are different in
  practice.
---

# Debug vs Release Build Mode

Basically, everything that runs with the binary that is built with `go run . build` is considered `release` mode which is ready for production deployment. Otherwise, it is `debug` mode.

Below is the snippet for working with the build mode:

```go
// Build is the current build type for the application, can be `debug` or 
// `release`. By default, it is appy.DebugBuild which is "debug".
// 
// Please take note that this value will be updated to `release` when running
// `go run . build` command.
var appy.Build = appy.DebugBuild

// IsDebugBuild indicates the current build is debug build which is meant for 
// local development.
func appy.IsDebugBuild() bool

// IsReleaseBuild indicates the current build is release build which is meant for 
// production deployment.
func appy.IsReleaseBuild() bool
```

Below are the main differences:

* assets/configs/mailers/translations/views are loaded from the filesystem for rapid development in `debug` mode.
* commands like `build` , `db:schema:dump` , `gen:migration` , `start` are not available in `debug` mode.
* messages that are logged via `app.Logger.Debug()` and `app.Logger.Debugf()` are only available in `debug` mode.
* `APPY_MASTER_KEY` can be read from either `configs/<APPY_ENV>.key` or the environment variable in `debug` mode whereas `release` mode strictly requires it to be the environment variable.
* The PWA is served via `webpack-dev-server` and proxied by the Go HTTP server in `debug` mode.

{% hint style="info" %}
Please refer to [https://bit.ly/3bqrB6M](https://bit.ly/3bqrB6M) for how logging overhead affects the latency performance.
{% endhint %}

