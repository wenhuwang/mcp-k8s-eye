package mcp

import (
	"slices"

	"github.com/mark3labs/mcp-go/server"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/k8s"
)

type Server struct {
	server *server.MCPServer
	k8s    *k8s.Kubernetes
}

func NewServer() (*Server, error) {
	s := &Server{
		server: server.NewMCPServer(
			"mcp-k8s-eye",
			"v0.0.1",
			server.WithResourceCapabilities(true, true),
			server.WithPromptCapabilities(true),
			server.WithLogging(),
		),
	}
	k8s, err := k8s.NewKubernetes()
	if err != nil {
		return nil, err
	}
	s.k8s = k8s

	s.server.AddTools(slices.Concat(
		s.initPods(),
	)...)

	return s, nil
}

func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.server)
}
