package utils

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func FetchLatestEvent(kubernetesClient *kubernetes.Clientset, namespace string, name string) (*v1.Event, error) {

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

func GetParent(client *kubernetes.Clientset, meta metav1.ObjectMeta) (string, bool) {
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

func IsErrorReason(reason string) bool {
	failureReasons := []string{
		"CrashLoopBackOff", "ImagePullBackOff", "CreateContainerConfigError", "PreCreateHookError", "CreateContainerError",
		"PreStartHookError", "RunContainerError", "ImageInspectError", "ErrImagePull", "ErrImageNeverPull", "InvalidImageName",
	}

	for _, r := range failureReasons {
		if r == reason {
			return true
		}
	}
	return false
}

func IsEvtErrorReason(reason string) bool {
	failureReasons := []string{
		"FailedCreatePodSandBox", "FailedMount",
	}

	for _, r := range failureReasons {
		if r == reason {
			return true
		}
	}
	return false
}
