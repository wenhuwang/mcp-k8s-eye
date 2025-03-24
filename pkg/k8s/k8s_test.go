package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewKubernetes(t *testing.T) {
	t.Run("Success case", func(t *testing.T) {
		// In normal situations, Kubernetes client should be created successfully
		k, err := NewKubernetes()

		// Assertions
		assert.NoError(t, err, "Should successfully create Kubernetes instance")
		assert.NotNil(t, k, "Kubernetes instance should not be nil")
		assert.NotNil(t, k.config, "config should not be nil")
		assert.NotNil(t, k.clientset, "clientset should not be nil")
		assert.NotNil(t, k.discoveryClient, "discoveryClient should not be nil")
		assert.NotNil(t, k.openapiSchema, "openapiSchema should not be nil")
	})
}
