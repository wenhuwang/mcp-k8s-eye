package k8s

import (
	openapi_v2 "github.com/google/gnostic/openapiv2"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Kubernetes struct {
	config                      *rest.Config
	clientset                   kubernetes.Interface
	discoveryClient             discovery.DiscoveryInterface
	dynamicClient               dynamic.Interface
	deferredDiscoveryRESTMapper *restmapper.DeferredDiscoveryRESTMapper
	openapiSchema               *openapi_v2.Document
	metricsClient               metricsclientset.Interface
}

// NewKubernetes creates a new Kubernetes client
func NewKubernetes() (*Kubernetes, error) {
	config, clientset, err := newK8SClient()
	if err != nil {
		return nil, err
	}

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	metricsClient, err := metricsclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Kubernetes{
		config:                      config,
		clientset:                   clientset,
		discoveryClient:             discoveryClient,
		dynamicClient:               dynamicClient,
		deferredDiscoveryRESTMapper: restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(discoveryClient)),
		openapiSchema:               &openapi_v2.Document{},
		metricsClient:               metricsClient,
	}, nil
}
