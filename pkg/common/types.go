package common

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
)

type PreAnalysis struct {
	Pod            v1.Pod
	FailureDetails []Failure
	Deployment     appsv1.Deployment
	ReplicaSet     appsv1.ReplicaSet
	Endpoint       v1.Endpoints
	Ingress        networkv1.Ingress
	StatefulSet    appsv1.StatefulSet
	Node           v1.Node
}
type Result struct {
	Kind         string    `json:"kind"`
	Name         string    `json:"name"`
	Error        []Failure `json:"error"`
	Details      string    `json:"details"`
	ParentObject string    `json:"parentObject"`
}

type Failure struct {
	Text          string
	KubernetesDoc string
	// Sensitive     []Sensitive
}

type Sensitive struct {
	Unmasked string
	Masked   string
}

type Request struct {
	Context       context.Context
	Namespace     string
	Name          string
	LabelSelector string
}
