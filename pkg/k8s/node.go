package k8s

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/utils"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k *Kubernetes) AnalyzeNode(ctx context.Context, name string) (string, error) {
	kind := "Node"
	nodes := make([]v1.Node, 0)
	if name == "" {
		nodeList, err := k.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return "", err
		}
		nodes = nodeList.Items
	} else {
		node, err := k.clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return "", err
		}
		nodes = append(nodes, *node)
	}
	var preAnalysis = map[string]common.PreAnalysis{}

	for _, node := range nodes {
		var failures []common.Failure
		for _, nodeCondition := range node.Status.Conditions {
			switch nodeCondition.Type {
			case v1.NodeReady:
				if nodeCondition.Status == v1.ConditionTrue {
					break
				}
				failures = addNodeConditionFailure(failures, node.Name, nodeCondition)
			default:
				if nodeCondition.Status != v1.ConditionFalse {
					failures = addNodeConditionFailure(failures, node.Name, nodeCondition)
				}
			}
		}

		if len(failures) > 0 {
			preAnalysis[node.Name] = common.PreAnalysis{
				Node:           node,
				FailureDetails: failures,
			}
		}
	}

	results := make([]common.Result, 0)
	for key, value := range preAnalysis {
		var result = common.Result{
			Kind:  kind,
			Name:  key,
			Error: value.FailureDetails,
		}
		parent, found := utils.GetParent(k.clientset, value.Node.ObjectMeta)
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

func addNodeConditionFailure(failures []common.Failure, nodeName string, nodeCondition v1.NodeCondition) []common.Failure {
	failures = append(failures, common.Failure{
		Text: fmt.Sprintf("%s condition type %s is %s, reason %s: %s", nodeName, nodeCondition.Type, nodeCondition.Status, nodeCondition.Reason, nodeCondition.Message),
	})
	return failures
}
