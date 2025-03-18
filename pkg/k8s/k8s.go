package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Kubernetes struct {
	config    *rest.Config
	clientset *kubernetes.Clientset
}

func NewKubernetes() (*Kubernetes, error) {
	config, clientset, err := newK8SClient()
	if err != nil {
		return nil, err
	}

	return &Kubernetes{
		config:    config,
		clientset: clientset,
	}, nil
}
