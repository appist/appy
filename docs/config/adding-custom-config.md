---
description: Covers how you can configure your application by adding custom config.
---

# Adding Custom Config

It is very common that an application would have its own configuration, e.g. database URI, 3rd party API credentials and etc. In order to accomodate this requirement while keeping the config management easier, `appy` combines both built-in and application config into 1 single config and place it under `app` package in your project module.

For example, we would like to integrate with [Facebook Graph API](https://developers.facebook.com/docs/graph-api/) and the API would require us to pass in `app_id` and `app_secret` to authenticate the request. And below is how we can configure the credentials:

#### Step 1: Add the credentials into the application config struct \(&lt;PROJECT\_NAME&gt;/pkg/app/config.go\).

```go
package app

type appConfig struct {
	AppName			string `env:"APP_NAME" envDefault:"<PROJECT_NAME>"`
	FBAppID 		string `env:"FB_APP_ID" envDefault:""`
	FBAppSecret string `env:"FB_APP_SECRET" envDefault:""`
}
```

#### Step 2: Encrypt the credentials and put it into the env config file.

```bash
// Encrypt the config value using the specific environment's secret key.
$ APPY_ENV=<ENV> go run . config:enc FB_APP_ID <FB_APP_ID_VALUE>
$ APPY_ENV=<ENV> go run . config:enc FB_APP_SECRET <FB_APP_SECRET_VALUE>
```

#### Step 3: Access the credentials in the codebase.

```go
import (
    "<PROJECT_NAME>/pkg/app"

    oauth2fb "golang.org/x/oauth2/facebook"
)

// Get Facebook access token.
conf := &oauth2.Config{
    ClientID:     app.Config.FBAppID,
    ClientSecret: app.Config.FBAppSecret,
    RedirectURL:  "...",
    Scopes:       []string{"email"},
    Endpoint:     oauth2fb.Endpoint,
}

token, err := conf.Exchange(oauth2.NoContext, "code")

...
```

Basically, this is how you can configure your application config, store securely and access it easily in your codebase.

{% hint style="info" %}
Note that appy's built-in config is embedded into `app.Config`, i.e. built-in config like `GQL_PLAYGROUND_ENABLED` is accessible via `app.Config.GQLPlaygroundEnabled` in the codebase. 
{% endhint %}

