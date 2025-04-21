package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAnalyzeNetworkPolicy(t *testing.T) {
	// Test a normal NetworkPolicy, should not have any issues
	clientset := fake.NewSimpleClientset(
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-policy",
				Namespace: "default",
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "test",
					},
				},
				Ingress: []networkingv1.NetworkPolicyIngressRule{
					{
						From: []networkingv1.NetworkPolicyPeer{
							{
								PodSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app": "frontend",
									},
								},
							},
						},
					},
				},
				PolicyTypes: []networkingv1.PolicyType{
					networkingv1.PolicyTypeIngress,
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
				Labels: map[string]string{
					"app": "test",
				},
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeNetworkPolicy(request)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "[]", result, "A normal NetworkPolicy should not have any issues")
}

func TestAnalyzeNetworkPolicyWithEmptyPodSelector(t *testing.T) {
	// Test a NetworkPolicy with an empty PodSelector
	clientset := fake.NewSimpleClientset(
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "empty-selector-policy",
				Namespace: "default",
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{}, // Empty selector
				Ingress: []networkingv1.NetworkPolicyIngressRule{
					{
						From: []networkingv1.NetworkPolicyPeer{
							{
								PodSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app": "frontend",
									},
								},
							},
						},
					},
				},
				PolicyTypes: []networkingv1.PolicyType{
					networkingv1.PolicyTypeIngress,
				},
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeNetworkPolicy(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "NetworkPolicy has empty pod selector", "Should detect empty Pod selector")
}

func TestAnalyzeNetworkPolicyWithNoMatchingPods(t *testing.T) {
	// Test a NetworkPolicy where the selector doesn't match any Pods
	clientset := fake.NewSimpleClientset(
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "no-matching-pods-policy",
				Namespace: "default",
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "non-existent",
					},
				},
				PolicyTypes: []networkingv1.PolicyType{
					networkingv1.PolicyTypeIngress,
				},
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeNetworkPolicy(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "NetworkPolicy has no matching pods", "Should detect no matching Pods")
}

func TestAnalyzeNetworkPolicyWithEmptyIngressRules(t *testing.T) {
	// Test a NetworkPolicy with no ingress rules but declared Ingress policy type
	clientset := fake.NewSimpleClientset(
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "empty-ingress-policy",
				Namespace: "default",
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				Ingress:     []networkingv1.NetworkPolicyIngressRule{}, // Empty ingress rules
				PolicyTypes: []networkingv1.PolicyType{
					networkingv1.PolicyTypeIngress,
				},
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeNetworkPolicy(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "NetworkPolicy will deny all ingress traffic", "Should detect empty ingress rules")
}

func TestAnalyzeNetworkPolicyWithEmptyEgressRules(t *testing.T) {
	// Test a NetworkPolicy with no egress rules but declared Egress policy type
	clientset := fake.NewSimpleClientset(
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "empty-egress-policy",
				Namespace: "default",
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				Egress:      []networkingv1.NetworkPolicyEgressRule{}, // Empty egress rules
				PolicyTypes: []networkingv1.PolicyType{
					networkingv1.PolicyTypeEgress,
				},
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeNetworkPolicy(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "NetworkPolicy will deny all egress traffic", "Should detect empty egress rules")
}

func TestAnalyzeNetworkPolicyNamespaceFiltering(t *testing.T) {
	// Test namespace filtering functionality
	clientset := fake.NewSimpleClientset(
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "policy-in-default",
				Namespace: "default",
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []networkingv1.PolicyType{
					networkingv1.PolicyTypeIngress,
				},
			},
		},
		&networkingv1.NetworkPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "policy-in-other",
				Namespace: "other",
			},
			Spec: networkingv1.NetworkPolicySpec{
				PodSelector: metav1.LabelSelector{},
				PolicyTypes: []networkingv1.PolicyType{
					networkingv1.PolicyTypeIngress,
				},
			},
		},
	)

	k := newTestKubernetes(clientset, nil)

	// Test default namespace
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeNetworkPolicy(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "default/policy-in-default", "Should contain policies in default namespace")
	assert.NotContains(t, result, "other/policy-in-other", "Should not contain policies in other namespace")

	// Test other namespace
	request = common.Request{
		Context:   context.Background(),
		Namespace: "other",
	}

	result, err = k.AnalyzeNetworkPolicy(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "other/policy-in-other", "Should contain policies in other namespace")
	assert.NotContains(t, result, "default/policy-in-default", "Should not contain policies in default namespace")
}
