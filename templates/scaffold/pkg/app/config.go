package app

type appConfig struct {
	AppName string `env:"APP_NAME" envDefault:"{{.Project.Name}}"`
}
