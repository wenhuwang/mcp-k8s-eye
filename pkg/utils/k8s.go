package utils

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	"sigs.k8s.io/yaml"
)

func FetchLatestEvent(kubernetesClient kubernetes.Interface, namespace string, name string) (*v1.Event, error) {

	// get the list of events
	events, err := kubernetesClient.CoreV1().Events(namespace).List(context.TODO(),
		metav1.ListOptions{
			FieldSelector: "involvedObject.name=" + name,
		})

	if err != nil {
		return nil, err
	}
	// find most recent event
	var latestEvent *v1.Event
	for _, event := range events.Items {
		if latestEvent == nil {
			// this is required, as a pointer to a loop variable would always yield the latest value in the range
			e := event
			latestEvent = &e
		}
		if event.LastTimestamp.After(latestEvent.LastTimestamp.Time) {
			// this is required, as a pointer to a loop variable would always yield the latest value in the range
			e := event
			latestEvent = &e
		}
	}
	return latestEvent, nil
}

func GetParent(client kubernetes.Interface, meta metav1.ObjectMeta) (string, bool) {
	if meta.OwnerReferences != nil {
		for _, owner := range meta.OwnerReferences {
			switch owner.Kind {
			case "ReplicaSet":
				rs, err := client.AppsV1().ReplicaSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if rs.OwnerReferences != nil {
					return GetParent(client, rs.ObjectMeta)
				}
				return "ReplicaSet/" + rs.Name, true

			case "Deployment":
				dep, err := client.AppsV1().Deployments(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if dep.OwnerReferences != nil {
					return GetParent(client, dep.ObjectMeta)
				}
				return "Deployment/" + dep.Name, true

			case "StatefulSet":
				sts, err := client.AppsV1().StatefulSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if sts.OwnerReferences != nil {
					return GetParent(client, sts.ObjectMeta)
				}
				return "StatefulSet/" + sts.Name, true

			case "DaemonSet":
				ds, err := client.AppsV1().DaemonSets(meta.Namespace).Get(context.Background(), owner.Name, metav1.GetOptions{})
				if err != nil {
					return "", false
				}
				if ds.OwnerReferences != nil {
					return GetParent(client, ds.ObjectMeta)
				}
				return "DaemonSet/" + ds.Name, true
			}
		}
	}
	return "", false
}

func Marshal(v any) (string, error) {
	switch t := v.(type) {
	case []unstructured.Unstructured:
		for i := range t {
			t[i].SetManagedFields(nil)
		}
	case *unstructured.Unstructured:
		t.SetManagedFields(nil)
	}
	res, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func NamespaceOrDefault(namespace string) string {
	if namespace == "" {
		return "default"
	}
	return namespace
}

var (
	supportedMetricsAPIVersions = []string{
		"v1beta1",
	}
)

func SupportedMetricsAPIVersionAvailable(discoveredAPIGroups *metav1.APIGroupList) bool {
	for _, discoveredAPIGroup := range discoveredAPIGroups.Groups {
		if discoveredAPIGroup.Name != metricsapi.GroupName {
			continue
		}
		for _, version := range discoveredAPIGroup.Versions {
			for _, supportedVersion := range supportedMetricsAPIVersions {
				if version.Version == supportedVersion {
					return true
				}
			}
		}
	}
	return false
}
