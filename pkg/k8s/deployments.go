package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DeploymentList lists all deployments in a namespace.
func (k *Kubernetes) DeploymentList(ctx context.Context, namespace string) (string, error) {
	deployList, err := k.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	cleaner := utils.NewResourceCleaner()
	cleaner.CleanList(deployList)

	return deployList.String(), nil
}

// DeploymentGet gets a deployment.
func (k *Kubernetes) DeploymentGet(ctx context.Context, namespace, name string) (string, error) {
	deploy, err := k.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	cleaner := utils.NewResourceCleaner()
	cleaner.Clean(deploy)

	return deploy.String(), nil
}

// DeploymentDelete deletes a deployment.
func (k *Kubernetes) DeploymentDelete(ctx context.Context, namespace, name string) (string, error) {
	err := k.clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return "", err
	}
	return "Deployment deleted successfully", nil
}

// DeploymentScale scales a deployment.
func (k *Kubernetes) DeploymentScale(ctx context.Context, namespace, name string, replicas int32) (string, error) {
	deploy, err := k.clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	deploy.Spec.Replicas = &replicas
	_, err = k.clientset.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Deployment %s/%s scaled to %d replicas", namespace, name, replicas), nil
}

// AnalyzeDeployments analyzes the deployments and returns a list of failures.
func (k *Kubernetes) AnalyzeDeployments(ctx context.Context, namespace string) (string, error) {
	kind := "Deployment"
	apiDoc := K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "apps",
			Version: "v1",
		},
		OpenapiSchema: k.openapiSchema,
	}

	deployList, err := k.clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, deploy := range deployList.Items {
		var failures []common.Failure

		if deploy.Status.AvailableReplicas < *deploy.Spec.Replicas {
			doc := apiDoc.GetApiDocV2("spec.replicas")
			failures = append(failures, common.Failure{
				Text:          fmt.Sprintf("Only %d/%d replicas available", deploy.Status.AvailableReplicas, *deploy.Spec.Replicas),
				KubernetesDoc: doc,
			})

			if len(failures) > 0 {
				preAnalysis[fmt.Sprintf("%s/%s", deploy.Namespace, deploy.Name)] = common.PreAnalysis{
					Deployment:     deploy,
					FailureDetails: failures,
				}
			}
		}
	}
	var results []common.Result
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
