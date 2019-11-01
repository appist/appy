package appy

type (
	Serverer interface {
		Hosts() []string
	}

	Server struct {
	}
)

func NewServer() *Server {
	return &Server{}
}

func (s Server) Hosts(config *Config) ([]string, error) {
	hosts := []string{}

	return hosts, nil
}
