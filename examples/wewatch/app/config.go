package app

type config struct {
	AppName string `env:"APP_NAME"`
}

// Config is the application config.
var Config *config

func init() {
	Config = &config{}
}
