package config

type app struct {
	AppName string `env:"APP_NAME"`
}

var App *app

func init() {
	App = &app{}
}
