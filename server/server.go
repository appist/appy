package server

type (
	// Server is mainly used to serve the HTTP requests.
	Server struct {
	}
)

// NewServer initializes Server instance.
func NewServer() *Server {
	return &Server{}
}
