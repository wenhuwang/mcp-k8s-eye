package k8s

import (
	"encoding/json"
	"fmt"

	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (k *Kubernetes) AnalyzeIngress(r common.Request) (string, error) {
	kind := "Ingress"
	apiDoc := K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "networking.k8s.io",
			Version: "v1",
		},
		OpenapiSchema: k.openapiSchema,
	}

	ingresses, err := k.clientset.NetworkingV1().Ingresses(r.Namespace).List(r.Context, metav1.ListOptions{
		LabelSelector: r.LabelSelector,
	})
	if err != nil {
		return "", err
	}

	var preAnalysis = map[string]common.PreAnalysis{}
	for _, ingress := range ingresses.Items {
		var failures []common.Failure

		// get ingress class
		ingressClassName := ingress.Spec.IngressClassName
		if ingressClassName == nil {
			ingClassValue := ingress.Annotations["kubernetes.io/ingress.class"]
			if ingClassValue == "" {
				doc := apiDoc.GetApiDocV2("spec.ingressClassName")
				failures = append(failures, common.Failure{
					Text:          "Ingress does not specify an ingress class",
					KubernetesDoc: doc,
				})
			} else {
				ingressClassName = &ingClassValue
			}
		}

		// check if ingressclass exists
		if ingressClassName != nil {
			_, err := k.clientset.NetworkingV1().IngressClasses().Get(r.Context, *ingressClassName, metav1.GetOptions{})
			if err != nil {
				doc := apiDoc.GetApiDocV2("spec.ingressClassName")
				failures = append(failures, common.Failure{
					Text:          fmt.Sprintf("Ingress uses the ingress class %s which does not exist", *ingressClassName),
					KubernetesDoc: doc,
				})
			}
		}

		// check if ingress uses a service that exists
		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}
			for _, path := range rule.HTTP.Paths {
				_, err := k.clientset.CoreV1().Services(ingress.Namespace).Get(r.Context, path.Backend.Service.Name, metav1.GetOptions{})
				if err != nil {
					doc := apiDoc.GetApiDocV2("spec.rules.http.paths.backend.service")
					failures = append(failures, common.Failure{
						Text: fmt.Sprintf(
							"Ingress uses the service %s/%s which does not exist",
							ingress.Namespace, path.Backend.Service.Name,
						),
						KubernetesDoc: doc,
					})
				}
			}
		}

		// check if ingress use a secret that exists
		for _, tls := range ingress.Spec.TLS {
			_, err := k.clientset.CoreV1().Secrets(ingress.Namespace).Get(r.Context, tls.SecretName, metav1.GetOptions{})
			if err != nil {
				doc := apiDoc.GetApiDocV2("spec.tls.secretName")
				failures = append(failures, common.Failure{
					Text: fmt.Sprintf(
						"Ingress uses the secret %s/%s which does not exist",
						ingress.Namespace, tls.SecretName,
					),
					KubernetesDoc: doc,
				})
			}
		}

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", ingress.Namespace, ingress.Name)] = common.PreAnalysis{
				Ingress:        ingress,
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
		parent, found := utils.GetParent(k.clientset, value.Ingress.ObjectMeta)
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
