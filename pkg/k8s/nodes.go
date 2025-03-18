package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kubernetes) NodeList(ctx context.Context) (string, error) {
	nodeList, err := k.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	return nodeList.String(), nil
}

func (k *Kubernetes) NodeGet(ctx context.Context, name string) (string, error) {
	node, err := k.clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return node.String(), nil
}
