package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

// ServiceList lists all services in a namespace.
func (k *Kubernetes) ServiceList(ctx context.Context, namespace string) (string, error) {
	svcList, err := k.clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	cleaner := utils.NewResourceCleaner()
	cleaner.CleanList(svcList)

	return svcList.String(), nil
}

// ServiceGet gets a service.
func (k *Kubernetes) ServiceGet(ctx context.Context, namespace, name string) (string, error) {
	svc, err := k.clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	cleaner := utils.NewResourceCleaner()
	cleaner.Clean(svc)

	return svc.String(), nil
}

// ServiceDelete deletes a service.
func (k *Kubernetes) ServiceDelete(ctx context.Context, namespace, name string) (string, error) {
	err := k.clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return "", err
	}
	return "Service deleted successfully", nil
}

// AnalyzeServices analyzes the services and returns a list of failures.
func (k *Kubernetes) AnalyzeServices(ctx context.Context, namespace string) (string, error) {
	kind := "Service"
	apiDoc := K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "",
			Version: "v1",
		},
		OpenapiSchema: k.openapiSchema,
	}

	epList, err := k.clientset.CoreV1().Endpoints(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, ep := range epList.Items {
		var failures []common.Failure

		if len(ep.Subsets) == 0 {
			if _, ok := ep.Annotations[resourcelock.LeaderElectionRecordAnnotationKey]; ok {
				continue
			}
			svc, err := k.clientset.CoreV1().Services(ep.Namespace).Get(ctx, ep.Name, metav1.GetOptions{})
			if err != nil {
				return "", err
			}

			for k, v := range svc.Spec.Selector {
				doc := apiDoc.GetApiDocV2("spec.selector")
				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("Service has no endpoints, unexpected label: %s=%s", k, v),
					KubernetesDoc: doc,
					Sensitive:     []common.Sensitive{},
				})
			}

		} else {
			count := 0
			pods := []string{}
			for _, subset := range ep.Subsets {
				apiDoc.Kind = "Endpoints"

				if len(subset.NotReadyAddresses) > 0 {
					for _, addr := range subset.NotReadyAddresses {
						count++
						pods = append(pods, addr.TargetRef.Kind+"/"+addr.TargetRef.Name)
					}
				}
			}

			if count > 0 {
				doc := apiDoc.GetApiDocV2("subsets.notReadyAddresses")
				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("Service has not ready endpoints, pods: %s, unexpected: %d", pods, count),
					KubernetesDoc: doc,
					Sensitive:     []common.Sensitive{},
				})
			}
		}

		// fetch event
		events, err := k.clientset.CoreV1().Events(ep.Namespace).List(ctx, metav1.ListOptions{
			FieldSelector: "involvedObject.name=" + ep.Name,
		})
		if err != nil {
			return "", err
		}

		for _, event := range events.Items {
			if event.Type != "Normal" {
				failures = append(failures, common.Failure{
					Text: fmt.Sprintf("Service %s/%s has event: %s", ep.Namespace, ep.Name, event.Message),
				})
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", ep.Namespace, ep.Name)] = common.PreAnalysis{
				Endpoint:       ep,
				FailureDetails: failures,
			}
		}
	}

	var results []common.Result
	for key, value := range preAnalysis {
		analysis := common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}

		parent, found := utils.GetParent(k.clientset, value.Endpoint.ObjectMeta)
		if found {
			analysis.ParentObject = parent
		}
		results = append(results, analysis)
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
