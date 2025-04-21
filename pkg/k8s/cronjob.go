package k8s

import (
	"encoding/json"
	"fmt"

	cron "github.com/robfig/cron/v3"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// AnalyzeCronJob analyzes the cronjobs and returns a list of failures.
func (k *Kubernetes) AnalyzeCronJob(r common.Request) (string, error) {
	kind := "CronJob"
	apiDoc := K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "batch",
			Version: "v1",
		},
		OpenapiSchema: k.openapiSchema,
	}
	cronJobList, err := k.clientset.BatchV1().CronJobs(r.Namespace).List(r.Context, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, cronjob := range cronJobList.Items {
		var failures []common.Failure

		// check if cronjob is suspended
		if cronjob.Spec.Suspend != nil && *cronjob.Spec.Suspend {
			doc := apiDoc.GetApiDocV2("spec.suspend")
			failures = append(failures, common.Failure{
				Text:          fmt.Sprintf("CronJob %s/%s is suspended", cronjob.Namespace, cronjob.Name),
				KubernetesDoc: doc,
			})
		} else {
			// check the schedule format
			if _, err := cron.ParseStandard(cronjob.Spec.Schedule); err != nil {
				doc := apiDoc.GetApiDocV2("spec.schedule")
				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("CronJob has an invalid schedule: %v", err),
					KubernetesDoc: doc,
				})
			}

			// check if cronjob has never been scheduled
			if cronjob.Status.LastScheduleTime == nil {
				failures = append(failures, common.Failure{
					Text: fmt.Sprint("CronJob has never been scheduled"),
				})
			}

			// check the starting deadline
			if cronjob.Spec.StartingDeadlineSeconds != nil {
				doc := apiDoc.GetApiDocV2("spec.startingDeadlineSeconds")
				if *cronjob.Spec.StartingDeadlineSeconds < 0 {
					failures = append(failures, common.Failure{
						Text:          fmt.Sprintf("CronJob has a negative starting deadline"),
						KubernetesDoc: doc,
					})
				}
			}
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", cronjob.Namespace, cronjob.Name)] = common.PreAnalysis{
				FailureDetails: failures,
			}
		}
	}

	results := make([]common.Result, 0)
	for key, value := range preAnalysis {
		result := common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}
		results = append(results, result)
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
