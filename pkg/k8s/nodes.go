package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeList lists all nodes.
func (k *Kubernetes) NodeList(ctx context.Context) (string, error) {
	nodeList, err := k.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	return nodeList.String(), nil
}

// NodeGet gets a node.
func (k *Kubernetes) NodeGet(ctx context.Context, name string) (string, error) {
	node, err := k.clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return node.String(), nil
}
