package config

type app struct {
	AppName string `env:"APP_NAME"`
}

// App is the application config.
var App *app

func init() {
	App = &app{}
}
