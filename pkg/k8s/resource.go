package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	metricsv1beta1api "k8s.io/metrics/pkg/apis/metrics/v1beta1"
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

func (k *Kubernetes) ResourceCreateOrUpdate(ctx context.Context, resource string) (string, error) {
	separator := regexp.MustCompile(`\r?\n---\r?\n`)
	resources := separator.Split(resource, -1)
	var unstructuredObjects []*unstructured.Unstructured
	for _, r := range resources {
		var obj *unstructured.Unstructured
		if err := yaml.NewYAMLToJSONDecoder(strings.NewReader(r)).Decode(&obj); err != nil {
			return "", err
		}
		unstructuredObjects = append(unstructuredObjects, obj)
	}

	return k.resourceCreateOrUpdate(ctx, unstructuredObjects)
}

func (k *Kubernetes) resourceCreateOrUpdate(ctx context.Context, resources []*unstructured.Unstructured) (string, error) {
	for _, obj := range resources {
		gvk := obj.GroupVersionKind()
		gvr, err := k.gvrFor(gvk)
		if err != nil {
			return "", err
		}
		namespace := obj.GetNamespace()
		if isNamespaced, err := k.isNamespaced(gvk); err == nil && isNamespaced {
			namespace = utils.NamespaceOrDefault(namespace)
		}
		_, err = k.dynamicClient.Resource(gvr).Namespace(namespace).Apply(ctx, obj.GetName(), obj, metav1.ApplyOptions{
			FieldManager: common.ProjectName,
		})
		if err != nil {
			return "", err
		}
		// Clear the cache to ensure the next operation is performed on the latest exposed APIs
		if gvk.Kind == "CustomResourceDefinition" {
			k.deferredDiscoveryRESTMapper.Reset()
		}
	}
	return fmt.Sprintf("All of the resource created/updated successfully"), nil
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

func (k *Kubernetes) WorkloadResourceUsage(r common.Request) (string, error) {
	var workloadMetrics = make(map[string][]metricsv1beta1api.PodMetrics, 0)
	switch r.WorkloadType {
	case "Deployment":
		if r.Name == "" {
			deployList, err := k.clientset.AppsV1().Deployments(r.Namespace).List(r.Context, metav1.ListOptions{
				LabelSelector: r.LabelSelector,
			})
			if err != nil {
				return "", err
			}
			for _, deploy := range deployList.Items {
				labelSelector := utils.MatchLabelsToLabelSelector(deploy.Spec.Selector.MatchLabels)
				metrics, err := utils.GetPodMetrics(r.Context, k.metricsClient, r.Namespace, "", labelSelector)
				if err != nil {
					return "", err
				}
				workloadMetrics[deploy.Name] = metrics.Items
			}
		} else {
			deploy, err := k.clientset.AppsV1().Deployments(r.Namespace).Get(r.Context, r.Name, metav1.GetOptions{})
			if err != nil {
				return "", err
			}
			labelSelector := utils.MatchLabelsToLabelSelector(deploy.Spec.Selector.MatchLabels)
			metrics, err := utils.GetPodMetrics(r.Context, k.metricsClient, r.Namespace, "", labelSelector)
			if err != nil {
				return "", err
			}
			workloadMetrics[r.Name] = metrics.Items
		}

	case "StatefulSet":
		if r.Name == "" {
			stsList, err := k.clientset.AppsV1().StatefulSets(r.Namespace).List(r.Context, metav1.ListOptions{
				LabelSelector: r.LabelSelector,
			})
			if err != nil {
				return "", err
			}
			for _, sts := range stsList.Items {
				labelSelector := utils.MatchLabelsToLabelSelector(sts.Spec.Selector.MatchLabels)
				metrics, err := utils.GetPodMetrics(r.Context, k.metricsClient, r.Namespace, "", labelSelector)
				if err != nil {
					return "", err
				}
				workloadMetrics[sts.Name] = metrics.Items
			}
		} else {
			sts, err := k.clientset.AppsV1().StatefulSets(r.Namespace).Get(r.Context, r.Name, metav1.GetOptions{})
			if err != nil {
				return "", err
			}
			labelSelector := utils.MatchLabelsToLabelSelector(sts.Spec.Selector.MatchLabels)
			metrics, err := utils.GetPodMetrics(r.Context, k.metricsClient, r.Namespace, "", labelSelector)
			if err != nil {
				return "", err
			}
			workloadMetrics[r.Name] = metrics.Items
		}

	case "DaemonSet":
		if r.Name == "" {
			dsList, err := k.clientset.AppsV1().DaemonSets(r.Namespace).List(r.Context, metav1.ListOptions{
				LabelSelector: r.LabelSelector,
			})
			if err != nil {
				return "", err
			}
			for _, ds := range dsList.Items {
				labelSelector := utils.MatchLabelsToLabelSelector(ds.Spec.Selector.MatchLabels)
				metrics, err := utils.GetPodMetrics(r.Context, k.metricsClient, r.Namespace, "", labelSelector)
				if err != nil {
					return "", err
				}
				workloadMetrics[ds.Name] = metrics.Items
			}
		} else {
			ds, err := k.clientset.AppsV1().DaemonSets(r.Namespace).Get(r.Context, r.Name, metav1.GetOptions{})
			if err != nil {
				return "", err
			}
			labelSelector := utils.MatchLabelsToLabelSelector(ds.Spec.Selector.MatchLabels)
			metrics, err := utils.GetPodMetrics(r.Context, k.metricsClient, r.Namespace, "", labelSelector)
			if err != nil {
				return "", err
			}
			workloadMetrics[r.Name] = metrics.Items
		}

	case "ReplicaSet":
		if r.Name == "" {
			rsList, err := k.clientset.AppsV1().ReplicaSets(r.Namespace).List(r.Context, metav1.ListOptions{
				LabelSelector: r.LabelSelector,
			})
			if err != nil {
				return "", err
			}
			for _, rs := range rsList.Items {
				labelSelector := utils.MatchLabelsToLabelSelector(rs.Spec.Selector.MatchLabels)
				metrics, err := utils.GetPodMetrics(r.Context, k.metricsClient, r.Namespace, "", labelSelector)
				if err != nil {
					return "", err
				}
				workloadMetrics[rs.Name] = metrics.Items
			}
		} else {
			rs, err := k.clientset.AppsV1().ReplicaSets(r.Namespace).Get(r.Context, r.Name, metav1.GetOptions{})
			if err != nil {
				return "", err
			}
			labelSelector := utils.MatchLabelsToLabelSelector(rs.Spec.Selector.MatchLabels)
			metrics, err := utils.GetPodMetrics(r.Context, k.metricsClient, r.Namespace, "", labelSelector)
			if err != nil {
				return "", err
			}
			workloadMetrics[r.Name] = metrics.Items
		}

	case "Pod":
		if r.Name == "" {
			metrics, err := utils.GetPodMetrics(r.Context, k.metricsClient, r.Namespace, "", r.LabelSelector)
			if err != nil {
				return "", err
			}
			for _, m := range metrics.Items {
				workloadMetrics[m.Name] = []metricsv1beta1api.PodMetrics{m}
			}
		} else {
			metrics, err := utils.GetPodMetrics(r.Context, k.metricsClient, r.Namespace, r.Name, r.LabelSelector)
			if err != nil {
				return "", err
			}
			workloadMetrics[r.Name] = metrics.Items
		}

	default:
		return "", fmt.Errorf("unknown workload type")
	}

	wfms := make([]*common.WorkloadFormattedMetrics, 0)
	for name, metrics := range workloadMetrics {
		var totalCPU, totalMemory int64
		for _, metric := range metrics {
			for _, c := range metric.Containers {
				totalCPU += c.Usage.Cpu().MilliValue()
				totalMemory += c.Usage.Memory().Value() / 1024 / 1024
			}
		}
		wfms = append(wfms, &common.WorkloadFormattedMetrics{
			Name:        name,
			Namespace:   r.Namespace,
			CPUUsage:    fmt.Sprintf("%dm", totalCPU),
			MemoryUsage: fmt.Sprintf("%dMi", totalMemory),
		})
	}

	jsonData, err := json.Marshal(wfms)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
