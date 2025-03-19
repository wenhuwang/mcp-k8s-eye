package k8s

import (
	openapi_v2 "github.com/google/gnostic/openapiv2"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Kubernetes struct {
	config        *rest.Config
	clientset     *kubernetes.Clientset
	openapiSchema *openapi_v2.Document
}

func NewKubernetes() (*Kubernetes, error) {
	config, clientset, err := newK8SClient()
	if err != nil {
		return nil, err
	}

	return &Kubernetes{
		config:        config,
		clientset:     clientset,
		openapiSchema: &openapi_v2.Document{},
	}, nil
}
