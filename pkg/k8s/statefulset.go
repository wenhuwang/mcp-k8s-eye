package k8s

import (
	"encoding/json"
	"fmt"

	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (k *Kubernetes) AnalyzeStatefulSet(r common.Request) (string, error) {
	kind := "StatefulSet"
	apiDoc := K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "apps",
			Version: "v1",
		},
		OpenapiSchema: k.openapiSchema,
	}

	stsList, err := k.clientset.AppsV1().StatefulSets(r.Namespace).List(r.Context, metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	var preAnalysis = map[string]common.PreAnalysis{}
	for _, sts := range stsList.Items {
		var failures []common.Failure

		svcName := sts.Spec.ServiceName
		if svcName != "" {
			_, err := k.clientset.CoreV1().Services(sts.Namespace).Get(r.Context, svcName, metav1.GetOptions{})
			if err != nil {
				doc := apiDoc.GetApiDocV2("spec.serviceName")
				failures = append(failures, common.Failure{
					Text: fmt.Sprintf(
						"StatefulSet uses the service %s/%s which does not exist",
						sts.Namespace, svcName,
					),
					KubernetesDoc: doc,
				})
			}
		}

		if len(sts.Spec.VolumeClaimTemplates) > 0 {
			for _, volumeClaimTemplate := range sts.Spec.VolumeClaimTemplates {
				_, err := k.clientset.CoreV1().PersistentVolumeClaims(sts.Namespace).Get(r.Context, volumeClaimTemplate.Name, metav1.GetOptions{})
				if err != nil {
					doc := apiDoc.GetApiDocV2("spec.volumeClaimTemplates")
					failures = append(failures, common.Failure{
						Text: fmt.Sprintf("StatefulSet uses the pvc %s/%s which does not exist",
							sts.Namespace, volumeClaimTemplate.Name,
						),
						KubernetesDoc: doc,
					})
				}
			}
		}

		if sts.Spec.Replicas != nil && *(sts.Spec.Replicas) != sts.Status.AvailableReplicas {
			for i := int32(0); i < *(sts.Spec.Replicas); i++ {
				podName := sts.Name + "-" + fmt.Sprint(i)
				pod, err := k.clientset.CoreV1().Pods(sts.Namespace).Get(r.Context, podName, metav1.GetOptions{})
				if err != nil {
					if errors.IsNotFound(err) && i == 0 {
						evt, err := utils.FetchLatestEvent(k.clientset, sts.Namespace, sts.Name)
						if err != nil || evt == nil || evt.Type == "Normal" {
							failures = append(failures, common.Failure{
								Text: fmt.Sprintf("StatefulSet has %d replicas, but only 0 pods are running", *(sts.Spec.Replicas)),
							})
							break
						}
						failures = append(failures, common.Failure{
							Text: evt.Message,
						})
					}
					break
				}
				if pod.Status.Phase != corev1.PodRunning {
					failures = append(failures, common.Failure{
						Text: fmt.Sprintf("StatefulSet pod %s/%s is not in Running state", sts.Namespace, podName),
					})
					break
				}
			}
		}
		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", sts.Namespace, sts.Name)] = common.PreAnalysis{
				StatefulSet:    sts,
				FailureDetails: failures,
			}
		}
	}

	results := make([]common.Result, 0)
	for key, value := range preAnalysis {
		analysis := common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}
		parent, found := utils.GetParent(k.clientset, value.StatefulSet.ObjectMeta)
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
