package k8s

import (
	"context"
	"fmt"

	"github.com/wenhuwang/mcp-k8s-eye/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (k *Kubernetes) ResourceList(ctx context.Context, kind, namespace string) (string, error) {
	kind = utils.Capitalize(kind)
	gv := utils.GetGroupVersionForKind(kind)
	gvk := gv.WithKind(kind)
	gvr, err := k.gvrFor(gvk)
	if err != nil {
		return "", err
	}

	isNamespaced, err := k.isNamespaced(gvk)
	if err != nil {
		return "", err
	}
	if isNamespaced {
		namespace = utils.NamespaceOrDefault(namespace)
	}

	resources, err := k.dynamicClient.Resource(gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	return utils.Marshal(resources.Items)
}

func (k *Kubernetes) ResourceGet(ctx context.Context, kind, namespace, name string) (string, error) {
	kind = utils.Capitalize(kind)
	gv := utils.GetGroupVersionForKind(kind)
	gvk := gv.WithKind(kind)
	gvr, err := k.gvrFor(gvk)
	if err != nil {
		return "", err
	}

	isNamespaced, err := k.isNamespaced(gvk)
	if err != nil {
		return "", err
	}

	if isNamespaced {
		namespace = utils.NamespaceOrDefault(namespace)
	}

	resource, err := k.dynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return utils.Marshal(resource)
}

func (k *Kubernetes) ResourceDelete(ctx context.Context, kind, namespace, name string) (string, error) {
	kind = utils.Capitalize(kind)
	gv := utils.GetGroupVersionForKind(kind)
	gvk := gv.WithKind(kind)
	gvr, err := k.gvrFor(gvk)
	if err != nil {
		return "", err
	}

	isNamespaced, err := k.isNamespaced(gvk)
	if err != nil {
		return "", err
	}

	if isNamespaced {
		namespace = utils.NamespaceOrDefault(namespace)
	}

	err = k.dynamicClient.Resource(gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Resource %s/%s deleted successfully", namespace, name), nil
}

func (k *Kubernetes) gvrFor(gvk schema.GroupVersionKind) (schema.GroupVersionResource, error) {
	mapping, err := k.deferredDiscoveryRESTMapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	return mapping.Resource, nil
}

func (k *Kubernetes) isNamespaced(gvk schema.GroupVersionKind) (bool, error) {
	apiResourceList, err := k.discoveryClient.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		return false, err
	}

	for _, apiResource := range apiResourceList.APIResources {
		if apiResource.Name == gvk.Kind {
			return apiResource.Namespaced, nil
		}
	}

	return false, nil
}
