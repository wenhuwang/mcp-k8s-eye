package k8s

import (
	"encoding/json"
	"fmt"

	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// AnalyzeValidatingWebhook analyzes ValidatingWebhookConfiguration resources and returns a list of failures
func (k *Kubernetes) AnalyzeValidatingWebhook(r common.Request) (string, error) {
	kind := "ValidatingWebhookConfiguration"
	apiDoc := K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "admissionregistration.k8s.io",
			Version: "v1",
		},
		OpenapiSchema: k.openapiSchema,
	}

	webhookList, err := k.clientset.AdmissionregistrationV1().ValidatingWebhookConfigurations().List(r.Context, metav1.ListOptions{
		LabelSelector: r.LabelSelector,
	})
	if err != nil {
		return "", err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, webhookConfig := range webhookList.Items {
		var failures []common.Failure

		// Check if the webhook has a valid service reference or URL
		for i, wh := range webhookConfig.Webhooks {
			// Check if webhook has a valid service reference
			if wh.ClientConfig.Service != nil {
				// Check if the service exists
				svc, err := k.clientset.CoreV1().Services(wh.ClientConfig.Service.Namespace).Get(
					r.Context, wh.ClientConfig.Service.Name, metav1.GetOptions{})
				if err != nil {
					doc := apiDoc.GetApiDocV2("webhooks.clientConfig.service")
					failures = append(failures, common.Failure{
						Text:          fmt.Sprintf("Webhook #%d (%s) references service which does not exist", i, wh.Name),
						KubernetesDoc: doc,
					})
				} else {
					// Check if the service port exists
					portExists := false
					for _, port := range svc.Spec.Ports {
						if wh.ClientConfig.Service.Port != nil && port.Port == *wh.ClientConfig.Service.Port {
							portExists = true
							break
						}
					}
					if wh.ClientConfig.Service.Port != nil && !portExists {
						doc := apiDoc.GetApiDocV2("webhooks.clientConfig.service.port")
						failures = append(failures, common.Failure{
							Text:          fmt.Sprintf("Webhook #%d (%s) references service port %d which does not exist in service", i, wh.Name, *wh.ClientConfig.Service.Port),
							KubernetesDoc: doc,
						})
					}

					if len(svc.Spec.Selector) == 0 {
						doc := apiDoc.GetApiDocV2("webhooks.clientConfig.service")
						failures = append(failures, common.Failure{
							Text:          fmt.Sprintf("Webhook #%d (%s) references service which does not have a selector", i, wh.Name),
							KubernetesDoc: doc,
						})
					}

					// Check if the pods of the service are running
					podList, err := k.clientset.CoreV1().Pods(wh.ClientConfig.Service.Namespace).List(r.Context, metav1.ListOptions{
						LabelSelector: utils.MapToString(svc.Spec.Selector),
					})
					if err != nil {
						return "", err
					}
					if len(podList.Items) == 0 {
						doc := apiDoc.GetApiDocV2("webhooks.clientConfig.service")
						failures = append(failures, common.Failure{
							Text:          fmt.Sprintf("Webhook #%d (%s) references service which does not have pods", i, wh.Name),
							KubernetesDoc: doc,
						})
					}

					// Check if the pods are running
					for _, pod := range podList.Items {
						if pod.Status.Phase != corev1.PodRunning {
							doc := apiDoc.GetApiDocV2("webhooks.clientConfig.service")
							failures = append(failures, common.Failure{
								Text:          fmt.Sprintf("Webhook #%d (%s) references service which have inactive pods", i, wh.Name),
								KubernetesDoc: doc,
							})
						}
					}
				}
			} else if wh.ClientConfig.URL == nil {
				doc := apiDoc.GetApiDocV2("webhooks.clientConfig")
				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("Webhook #%d (%s) has neither service reference nor URL", i, wh.Name),
					KubernetesDoc: doc,
				})
			}

			// Check if the CA bundle is empty
			if len(wh.ClientConfig.CABundle) == 0 {
				doc := apiDoc.GetApiDocV2("webhooks.clientConfig.caBundle")
				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("Webhook #%d (%s) has empty CA bundle", i, wh.Name),
					KubernetesDoc: doc,
				})
			}

			// Check if the webhook has rules
			if wh.Rules == nil || len(wh.Rules) == 0 {
				doc := apiDoc.GetApiDocV2("webhooks.rules")
				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("Webhook #%d (%s) has no rules", i, wh.Name),
					KubernetesDoc: doc,
				})
			}
		}

		if len(failures) > 0 {
			preAnalysis[webhookConfig.Name] = common.PreAnalysis{
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

// AnalyzeMutatingWebhook analyzes MutatingWebhookConfiguration resources and returns a list of failures
func (k *Kubernetes) AnalyzeMutatingWebhook(r common.Request) (string, error) {
	kind := "MutatingWebhookConfiguration"
	apiDoc := K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "admissionregistration.k8s.io",
			Version: "v1",
		},
		OpenapiSchema: k.openapiSchema,
	}

	webhookList, err := k.clientset.AdmissionregistrationV1().MutatingWebhookConfigurations().List(r.Context, metav1.ListOptions{
		LabelSelector: r.LabelSelector,
	})
	if err != nil {
		return "", err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, webhook := range webhookList.Items {
		var failures []common.Failure

		// Check if the webhook has a valid service reference or URL
		for i, wh := range webhook.Webhooks {
			// Check if webhook has a valid service reference
			if wh.ClientConfig.Service != nil {
				// Check if the service exists
				svc, err := k.clientset.CoreV1().Services(wh.ClientConfig.Service.Namespace).Get(
					r.Context, wh.ClientConfig.Service.Name, metav1.GetOptions{})
				if err != nil {
					doc := apiDoc.GetApiDocV2("webhooks.clientConfig.service")
					failures = append(failures, common.Failure{
						Text:          fmt.Sprintf("Webhook #%d (%s) references service %s/%s which does not exist", i, wh.Name, wh.ClientConfig.Service.Namespace, wh.ClientConfig.Service.Name),
						KubernetesDoc: doc,
					})
				} else {
					// Check if the service port exists
					portExists := false
					for _, port := range svc.Spec.Ports {
						if wh.ClientConfig.Service.Port != nil && port.Port == *wh.ClientConfig.Service.Port {
							portExists = true
							break
						}
					}
					if wh.ClientConfig.Service.Port != nil && !portExists {
						doc := apiDoc.GetApiDocV2("webhooks.clientConfig.service.port")
						failures = append(failures, common.Failure{
							Text:          fmt.Sprintf("Webhook #%d (%s) references service port %d which does not exist in service %s/%s", i, wh.Name, *wh.ClientConfig.Service.Port, wh.ClientConfig.Service.Namespace, wh.ClientConfig.Service.Name),
							KubernetesDoc: doc,
						})
					}

					if len(svc.Spec.Selector) == 0 {
						doc := apiDoc.GetApiDocV2("webhooks.clientConfig.service")
						failures = append(failures, common.Failure{
							Text:          fmt.Sprintf("Webhook #%d (%s) references service which does not have a selector", i, wh.Name),
							KubernetesDoc: doc,
						})
					}
					// Check if the pods of the service are running
					podList, err := k.clientset.CoreV1().Pods(wh.ClientConfig.Service.Namespace).List(r.Context, metav1.ListOptions{
						LabelSelector: utils.MapToString(svc.Spec.Selector),
					})
					if err != nil {
						return "", err
					}
					if len(podList.Items) == 0 {
						doc := apiDoc.GetApiDocV2("webhooks.clientConfig.service")
						failures = append(failures, common.Failure{
							Text:          fmt.Sprintf("Webhook #%d (%s) references service which does not have pods", i, wh.Name),
							KubernetesDoc: doc,
						})
					}
					// Check if the pods are running
					for _, pod := range podList.Items {
						if pod.Status.Phase != corev1.PodRunning {
							doc := apiDoc.GetApiDocV2("webhooks.clientConfig.service")
							failures = append(failures, common.Failure{
								Text:          fmt.Sprintf("Webhook #%d (%s) references service which have inactive pods", i, wh.Name),
								KubernetesDoc: doc,
							})
						}
					}
				}
			} else if wh.ClientConfig.URL == nil {
				doc := apiDoc.GetApiDocV2("webhooks.clientConfig")
				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("Webhook #%d (%s) has neither service reference nor URL", i, wh.Name),
					KubernetesDoc: doc,
				})
			}

			// Check if the CA bundle is empty
			if len(wh.ClientConfig.CABundle) == 0 {
				doc := apiDoc.GetApiDocV2("webhooks.clientConfig.caBundle")
				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("Webhook #%d (%s) has empty CA bundle", i, wh.Name),
					KubernetesDoc: doc,
				})
			}

			// Check if the webhook has rules
			if wh.Rules == nil || len(wh.Rules) == 0 {
				doc := apiDoc.GetApiDocV2("webhooks.rules")
				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("Webhook #%d (%s) has no rules", i, wh.Name),
					KubernetesDoc: doc,
				})
			}
		}
		if len(failures) > 0 {
			preAnalysis[webhook.Name] = common.PreAnalysis{
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
