package cmd

import (
	"fmt"
	"net"
	"os"
	"path"
	"runtime"

	"github.com/appist/appy/support"
	"github.com/spf13/cobra"
)

// VERSION is the current version of appy.
const VERSION = "0.1.0"

var (
	config           *support.ConfigT
	logger           *support.LoggerT
	root             *cobra.Command
	reservedCmdNames = map[string]bool{}
)

func init() {
	config = support.Config
	logger = support.Logger
	root = NewCommand()
}

// NewCommand initializes the root command instance.
func NewCommand() *cobra.Command {
	cmdName := path.Base(os.Args[0])
	if cmdName == "main" {
		wd, err := os.Getwd()
		if err != nil {
			logger.Fatal(err)
		}

		cmdName = path.Base(wd)
	}

	return &cobra.Command{
		Use:     cmdName,
		Short:   "An opinionated productive web framework that helps scaling business easier.",
		Version: VERSION,
	}
}

// AddCommand adds a custom command.
func AddCommand(command *cobra.Command) {
	if _, ok := reservedCmdNames[command.Name()]; ok {
		logger.Fatalf("'%s' command name is reserved, please update the command name.", command.Name())
	}

	root.AddCommand(command)
}

// Run executes the root command.
func Run() {
	root.Execute()
}

func checkSSLCerts() {
	if config.HTTPSSLEnabled == true {
		if _, err := os.Stat(config.HTTPSSLCertPath + "/cert.pem"); os.IsNotExist(err) {
			logger.Fatal("HTTP_SSL_ENABLED is set to true without SSL certs, please generate using `go run . ssl:setup` first.")
		}

		if _, err := os.Stat(config.HTTPSSLCertPath + "/key.pem"); os.IsNotExist(err) {
			logger.Fatal("HTTP_SSL_ENABLED is set to true without SSL certs, please generate using `go run . ssl:setup` first.")
		}
	}
}

func getIPHosts() []string {
	var hosts = []string{config.HTTPHost}

	if config.HTTPHost != "localhost" {
		hosts = append(hosts, "localhost")
	}

	addresses, err := net.InterfaceAddrs()
	if err != nil {
		logger.Fatal(err)
	}

	for _, address := range addresses {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			host := ipnet.IP.To4()
			if host != nil {
				hosts = append(hosts, host.String())
			}
		}
	}

	return hosts
}

func logServerInfo(sslEnabled bool, httpPort, httpSSLPort string) {
	logger.Infof("* Version %s (%s), build: %s", VERSION, runtime.Version(), support.Build)
	logger.Infof("* Environment: %s", config.AppyEnv)
	logger.Infof("* Environment Config: %s", support.DotenvPath)

	hosts := getIPHosts()
	host := fmt.Sprintf("http://%s:%s", hosts[0], httpPort)

	if sslEnabled == true {
		host = fmt.Sprintf("https://%s:%s", hosts[0], httpSSLPort)
	}

	logger.Infof("* Listening on %s", host)
}
