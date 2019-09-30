package http

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"

	"github.com/appist/appy/support"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ServerT is a type that contains GRPC/HTTP servers and the gin router in it.
type ServerT struct {
	Config *support.ConfigT
	GRPC   *grpc.Server
	HTTP   *http.Server
	Router *RouterT
}

// NewServer returns the server instance which contains GRPC/HTTP servers and the gin router in it.
func NewServer(c *support.ConfigT) *ServerT {
	r := newRouter(c)
	s := &http.Server{
		Addr:              c.HTTPHost + ":" + c.HTTPPort,
		Handler:           r,
		MaxHeaderBytes:    c.HTTPMaxHeaderBytes,
		ReadTimeout:       c.HTTPReadTimeout,
		ReadHeaderTimeout: c.HTTPReadHeaderTimeout,
		WriteTimeout:      c.HTTPWriteTimeout,
		IdleTimeout:       c.HTTPIdleTimeout,
	}
	s.ErrorLog = zap.NewStdLog(support.Logger.Desugar())

	if c.HTTPSSLEnabled == true {
		s.Addr = c.HTTPHost + ":" + c.HTTPSSLPort
	}

	return &ServerT{
		Config: c,
		GRPC:   nil, // to be implemented
		HTTP:   s,
		Router: r,
	}
}

// CheckSSLCerts checks if `./tmp/ssl` exists and contains the locally trusted SSL certificates.
func (s *ServerT) CheckSSLCerts() {
	if s.Config.HTTPSSLEnabled == true {
		if _, err := os.Stat(s.Config.HTTPSSLCertPath + "/cert.pem"); os.IsNotExist(err) {
			support.Logger.Fatal("HTTP_SSL_ENABLED is set to true without SSL certs, please generate using `go run . ssl:setup` first.")
		}

		if _, err := os.Stat(s.Config.HTTPSSLCertPath + "/key.pem"); os.IsNotExist(err) {
			support.Logger.Fatal("HTTP_SSL_ENABLED is set to true without SSL certs, please generate using `go run . ssl:setup` first.")
		}
	}
}

// Hosts returns the server hosts list.
func (s *ServerT) Hosts() []string {
	var hosts = []string{s.Config.HTTPHost}

	if s.Config.HTTPHost != "localhost" {
		hosts = append(hosts, "localhost")
	}

	addresses, err := net.InterfaceAddrs()
	if err != nil {
		support.Logger.Fatal(err)
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

// Routes returns all the routes including those in middlewares.
func (s *ServerT) Routes() []RouteInfoT {
	routes := s.Router.Routes()

	if s.Config.HTTPHealthCheckURL != "" {
		routes = append(routes, RouteInfoT{
			Method:      "GET",
			Path:        s.Config.HTTPHealthCheckURL,
			Handler:     "",
			HandlerFunc: nil,
		})
	}

	return routes
}

// PrintInfo prints the server info.
func (s *ServerT) PrintInfo() {
	support.Logger.Infof("* Version %s (%s), build: %s", support.VERSION, runtime.Version(), support.Build)
	support.Logger.Infof("* Environment: %s", s.Config.AppyEnv)
	support.Logger.Infof("* Environment Config: %s", support.DotenvPath)

	hosts := s.Hosts()
	host := fmt.Sprintf("http://%s:%s", hosts[0], s.Config.HTTPPort)

	if s.Config.HTTPSSLEnabled == true {
		host = fmt.Sprintf("https://%s:%s", hosts[0], s.Config.HTTPSSLPort)
	}

	support.Logger.Infof("* Listening on %s", host)
}
