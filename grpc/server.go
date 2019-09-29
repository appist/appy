package grpc // TODO: Implement full-fledged features set to support GRPC development.

import (
	"google.golang.org/grpc"
)

// ServerT is a high level gRPC server that provides more functionalities to serve gRPC requests.
type ServerT struct {
	server *grpc.Server
}

// NewServer returns a gRPC server instance.
func NewServer() *ServerT {
	s := &ServerT{}

	return s
}
