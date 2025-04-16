package k8s

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAnalyzeNodeWithoutCondition(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
			},
		})
	k := newTestKubernetes(clientset, nil)
	result, err := k.AnalyzeNode(context.Background(), "node1")
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "[]")
}

func TestAnalyzeNodeWithCondition(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&v1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
			},
			Status: v1.NodeStatus{
				Conditions: []v1.NodeCondition{
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionFalse,
					},
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionUnknown,
					},
					{
						Type:   v1.NodeReady,
						Status: v1.ConditionTrue,
					},
					{
						Type:   v1.NodeMemoryPressure,
						Status: v1.ConditionTrue,
					},
					{
						Type:   v1.NodeDiskPressure,
						Status: v1.ConditionTrue,
					},
					{
						Type:   v1.NodePIDPressure,
						Status: v1.ConditionTrue,
					},
					{
						Type:   v1.NodeNetworkUnavailable,
						Status: v1.ConditionTrue,
					},
					{
						Type:   v1.NodeMemoryPressure,
						Status: v1.ConditionUnknown,
					},
					{
						Type:   v1.NodeDiskPressure,
						Status: v1.ConditionUnknown,
					},
					{
						Type:   v1.NodePIDPressure,
						Status: v1.ConditionUnknown,
					},
					{
						Type:   v1.NodeNetworkUnavailable,
						Status: v1.ConditionUnknown,
					},
					{
						Type:   v1.NodeMemoryPressure,
						Status: v1.ConditionFalse,
					},
					{
						Type:   v1.NodeDiskPressure,
						Status: v1.ConditionFalse,
					},
					{
						Type:   v1.NodePIDPressure,
						Status: v1.ConditionFalse,
					},
					{
						Type:   v1.NodeNetworkUnavailable,
						Status: v1.ConditionFalse,
					},
				},
			},
		},
	)
	k := newTestKubernetes(clientset, nil)
	result, err := k.AnalyzeNode(context.Background(), "node1")
	if err != nil {
		t.Error(err)
	}
	count := strings.Count(result, "node1")
	assert.Equal(t, 11, count)
}
