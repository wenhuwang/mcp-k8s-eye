package mcp

import (
	"testing"
)

func TestNewServer(t *testing.T) {
	t.Run("Can instantiate", func(t *testing.T) {
		server, err := NewServer("test-server", "1.0.0")
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		if server == nil {
			t.Error("Server should not be nil")
		}
		if server.server == nil {
			t.Error("MCPServer should not be nil")
		}
		if server.k8s == nil {
			t.Error("Kubernetes config should not be nil")
		}
	})
}
