package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
)

// PodLogs returns the logs of a pod.
func (k *Kubernetes) PodLogs(ctx context.Context, namespace, name string) (string, error) {
	tailLines := int64(200)
	req := k.clientset.CoreV1().Pods(namespace).GetLogs(name, &v1.PodLogOptions{
		TailLines: &tailLines,
	})
	res := req.Do(ctx)
	if res.Error() != nil {
		return "", res.Error()
	}

	rawData, err := res.Raw()
	if err != nil {
		return "", err
	}
	return string(rawData), nil
}

// PodExec executes a command in a pod and returns the output.
func (k *Kubernetes) PodExec(ctx context.Context, namespace, name, command string) (string, error) {
	req := k.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(name).
		Namespace(namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Command: strings.Split(command, " "),
		Stdin:   false,
		Stdout:  true,
		Stderr:  true,
		TTY:     false,
	}, metav1.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(k.config, "POST", req.URL())
	if err != nil {
		return "", err
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if err != nil {
		return "", err
	}

	if stderr.Len() > 0 {
		return "", fmt.Errorf(stderr.String())
	}

	return stdout.String(), nil
}

// AnalyzePods analyzes the pods and returns a list of failures.
func (k *Kubernetes) AnalyzePod(ctx context.Context, namespace string) (string, error) {
	podList, err := k.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var preAnalysis = map[string]common.PreAnalysis{}

	for _, pod := range podList.Items {
		var failures []common.Failure

		// Check for pending pods
		if pod.Status.Phase == "Pending" {
			// Check through container status to check for crashes
			for _, containerStatus := range pod.Status.Conditions {
				if containerStatus.Type == v1.PodScheduled && containerStatus.Reason == "Unschedulable" {
					if containerStatus.Message != "" {
						failures = append(failures, common.Failure{
							Text: containerStatus.Message,
						})
					}
				}
			}
		}

		// Check for errors in the init containers.
		failures = append(failures, k.analyzeContainerStatusFailures(pod.Status.InitContainerStatuses, pod.Name, pod.Namespace, string(pod.Status.Phase))...)

		// Check for errors in containers.
		failures = append(failures, k.analyzeContainerStatusFailures(pod.Status.ContainerStatuses, pod.Name, pod.Namespace, string(pod.Status.Phase))...)

		if len(failures) > 0 {
			preAnalysis[fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)] = common.PreAnalysis{
				Pod:            pod,
				FailureDetails: failures,
			}
		}
	}

	results := make([]common.Result, 0)
	for key, value := range preAnalysis {
		result := common.Result{
			Kind:  "Pod",
			Name:  key,
			Error: value.FailureDetails,
		}
		parent, found := utils.GetParent(k.clientset, value.Pod.ObjectMeta)
		if found {
			result.ParentObject = parent
		}
		results = append(results, result)
	}

	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// analyzeContainerStatusFailures analyzes the container statuses and returns a list of failures.
func (k *Kubernetes) analyzeContainerStatusFailures(statuses []v1.ContainerStatus, name string, namespace string, statusPhase string) []common.Failure {
	var failures []common.Failure

	// Check through container status to check for crashes or unready
	for _, containerStatus := range statuses {
		if containerStatus.State.Waiting != nil {
			if containerStatus.State.Waiting.Reason == "ContainerCreating" && statusPhase == "Pending" {
				// This represents a container that is still being created or blocked due to conditions such as OOMKilled
				// parse the event log and append details
				evt, err := utils.FetchLatestEvent(k.clientset, namespace, name)
				if err != nil || evt == nil {
					continue
				}
				if utils.IsEvtErrorReason(evt.Reason) && evt.Message != "" {
					failures = append(failures, common.Failure{
						Text: evt.Message,
					})
				}
			} else if containerStatus.State.Waiting.Reason == "CrashLoopBackOff" && containerStatus.LastTerminationState.Terminated != nil {
				// This represents container that is in CrashLoopBackOff state due to conditions such as OOMKilled
				failures = append(failures, common.Failure{
					Text: fmt.Sprintf("the last termination reason is %s container=%s pod=%s", containerStatus.LastTerminationState.Terminated.Reason, containerStatus.Name, name),
				})
			} else if utils.IsErrorReason(containerStatus.State.Waiting.Reason) && containerStatus.State.Waiting.Message != "" {
				failures = append(failures, common.Failure{
					Text: containerStatus.State.Waiting.Message,
				})
			}
		} else {
			// when pod is Running but its ReadinessProbe fails
			if !containerStatus.Ready && statusPhase == "Running" {
				// parse the event log and append details
				evt, err := utils.FetchLatestEvent(k.clientset, namespace, name)
				if err != nil || evt == nil {
					continue
				}
				if evt.Reason == "Unhealthy" && evt.Message != "" {
					failures = append(failures, common.Failure{
						Text: evt.Message,
					})
				}
			}
		}
	}

	return failures
}
