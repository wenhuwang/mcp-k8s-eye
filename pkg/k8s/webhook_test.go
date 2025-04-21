package k8s

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAnalyzeValidatingWebhook(t *testing.T) {
	// Test a normal ValidatingWebhookConfiguration, should not have any issues
	port := int32(443)
	clientset := fake.NewSimpleClientset(
		&admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-validating-webhook",
			},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{
				{
					Name: "test-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "default",
							Name:      "webhook-service",
							Port:      &port,
						},
						CABundle: []byte("test-ca-bundle"),
					},
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{
								admissionregistrationv1.Create,
							},
							Rule: admissionregistrationv1.Rule{
								APIGroups:   []string{"apps"},
								APIVersions: []string{"v1"},
								Resources:   []string{"deployments"},
							},
						},
					},
				},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": "webhook",
				},
				Ports: []corev1.ServicePort{
					{
						Port: 443,
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook-pod",
				Namespace: "default",
				Labels: map[string]string{
					"app": "webhook",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeValidatingWebhook(request)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "[]", result, "A normal ValidatingWebhookConfiguration should not have any issues")
}

func TestAnalyzeValidatingWebhookWithNonExistentService(t *testing.T) {
	// Test a ValidatingWebhookConfiguration with a non-existent service
	port := int32(443)
	clientset := fake.NewSimpleClientset(
		&admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-validating-webhook",
			},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{
				{
					Name: "test-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "default",
							Name:      "non-existent-service",
							Port:      &port,
						},
						CABundle: []byte("test-ca-bundle"),
					},
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{
								admissionregistrationv1.Create,
							},
							Rule: admissionregistrationv1.Rule{
								APIGroups:   []string{"apps"},
								APIVersions: []string{"v1"},
								Resources:   []string{"deployments"},
							},
						},
					},
				},
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeValidatingWebhook(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "references service which does not exist", "Should detect non-existent service")
}

func TestAnalyzeValidatingWebhookWithNonExistentServicePort(t *testing.T) {
	// Test a ValidatingWebhookConfiguration with a non-existent service port
	port := int32(8443)
	clientset := fake.NewSimpleClientset(
		&admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-validating-webhook",
			},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{
				{
					Name: "test-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "default",
							Name:      "webhook-service",
							Port:      &port,
						},
						CABundle: []byte("test-ca-bundle"),
					},
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{
								admissionregistrationv1.Create,
							},
							Rule: admissionregistrationv1.Rule{
								APIGroups:   []string{"apps"},
								APIVersions: []string{"v1"},
								Resources:   []string{"deployments"},
							},
						},
					},
				},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": "webhook",
				},
				Ports: []corev1.ServicePort{
					{
						Port: 443,
					},
				},
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeValidatingWebhook(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, fmt.Sprintf("references service port %d which does not exist in service", 8443), "Should detect non-existent service port")
}

func TestAnalyzeValidatingWebhookWithNoSelector(t *testing.T) {
	// Test a ValidatingWebhookConfiguration with a service that has no selector
	port := int32(443)
	clientset := fake.NewSimpleClientset(
		&admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-validating-webhook",
			},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{
				{
					Name: "test-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "default",
							Name:      "webhook-service",
							Port:      &port,
						},
						CABundle: []byte("test-ca-bundle"),
					},
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{
								admissionregistrationv1.Create,
							},
							Rule: admissionregistrationv1.Rule{
								APIGroups:   []string{"apps"},
								APIVersions: []string{"v1"},
								Resources:   []string{"deployments"},
							},
						},
					},
				},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{}, // Empty selector
				Ports: []corev1.ServicePort{
					{
						Port: 443,
					},
				},
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeValidatingWebhook(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "references service which does not have a selector", "Should detect service with no selector")
}

func TestAnalyzeValidatingWebhookWithNoPods(t *testing.T) {
	// Test a ValidatingWebhookConfiguration with a service that has no pods
	port := int32(443)
	clientset := fake.NewSimpleClientset(
		&admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-validating-webhook",
			},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{
				{
					Name: "test-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "default",
							Name:      "webhook-service",
							Port:      &port,
						},
						CABundle: []byte("test-ca-bundle"),
					},
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{
								admissionregistrationv1.Create,
							},
							Rule: admissionregistrationv1.Rule{
								APIGroups:   []string{"apps"},
								APIVersions: []string{"v1"},
								Resources:   []string{"deployments"},
							},
						},
					},
				},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": "webhook",
				},
				Ports: []corev1.ServicePort{
					{
						Port: 443,
					},
				},
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeValidatingWebhook(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "references service which does not have pods", "Should detect service with no pods")
}

func TestAnalyzeValidatingWebhookWithInactivePods(t *testing.T) {
	// Test a ValidatingWebhookConfiguration with a service that has inactive pods
	port := int32(443)
	clientset := fake.NewSimpleClientset(
		&admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-validating-webhook",
			},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{
				{
					Name: "test-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "default",
							Name:      "webhook-service",
							Port:      &port,
						},
						CABundle: []byte("test-ca-bundle"),
					},
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{
								admissionregistrationv1.Create,
							},
							Rule: admissionregistrationv1.Rule{
								APIGroups:   []string{"apps"},
								APIVersions: []string{"v1"},
								Resources:   []string{"deployments"},
							},
						},
					},
				},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": "webhook",
				},
				Ports: []corev1.ServicePort{
					{
						Port: 443,
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook-pod",
				Namespace: "default",
				Labels: map[string]string{
					"app": "webhook",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending, // Inactive pod
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeValidatingWebhook(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "references service which have inactive pods", "Should detect service with inactive pods")
}

func TestAnalyzeValidatingWebhookWithNoClientConfig(t *testing.T) {
	// Test a ValidatingWebhookConfiguration with no service or URL
	clientset := fake.NewSimpleClientset(
		&admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-validating-webhook",
			},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{
				{
					Name: "test-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						// No service or URL
						CABundle: []byte("test-ca-bundle"),
					},
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{
								admissionregistrationv1.Create,
							},
							Rule: admissionregistrationv1.Rule{
								APIGroups:   []string{"apps"},
								APIVersions: []string{"v1"},
								Resources:   []string{"deployments"},
							},
						},
					},
				},
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeValidatingWebhook(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "has neither service reference nor URL", "Should detect webhook with no service or URL")
}

func TestAnalyzeValidatingWebhookWithEmptyCABundle(t *testing.T) {
	// Test a ValidatingWebhookConfiguration with an empty CA bundle
	port := int32(443)
	clientset := fake.NewSimpleClientset(
		&admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-validating-webhook",
			},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{
				{
					Name: "test-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "default",
							Name:      "webhook-service",
							Port:      &port,
						},
						CABundle: []byte{}, // Empty CA bundle
					},
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{
								admissionregistrationv1.Create,
							},
							Rule: admissionregistrationv1.Rule{
								APIGroups:   []string{"apps"},
								APIVersions: []string{"v1"},
								Resources:   []string{"deployments"},
							},
						},
					},
				},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": "webhook",
				},
				Ports: []corev1.ServicePort{
					{
						Port: 443,
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook-pod",
				Namespace: "default",
				Labels: map[string]string{
					"app": "webhook",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeValidatingWebhook(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "has empty CA bundle", "Should detect webhook with empty CA bundle")
}

func TestAnalyzeValidatingWebhookWithNoRules(t *testing.T) {
	// Test a ValidatingWebhookConfiguration with no rules
	port := int32(443)
	clientset := fake.NewSimpleClientset(
		&admissionregistrationv1.ValidatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-validating-webhook",
			},
			Webhooks: []admissionregistrationv1.ValidatingWebhook{
				{
					Name: "test-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "default",
							Name:      "webhook-service",
							Port:      &port,
						},
						CABundle: []byte("test-ca-bundle"),
					},
					Rules: []admissionregistrationv1.RuleWithOperations{}, // No rules
				},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": "webhook",
				},
				Ports: []corev1.ServicePort{
					{
						Port: 443,
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook-pod",
				Namespace: "default",
				Labels: map[string]string{
					"app": "webhook",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeValidatingWebhook(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "has no rules", "Should detect webhook with no rules")
}

func TestAnalyzeMutatingWebhook(t *testing.T) {
	// Test a normal MutatingWebhookConfiguration, should not have any issues
	port := int32(443)
	clientset := fake.NewSimpleClientset(
		&admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-mutating-webhook",
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{
				{
					Name: "test-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "default",
							Name:      "webhook-service",
							Port:      &port,
						},
						CABundle: []byte("test-ca-bundle"),
					},
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{
								admissionregistrationv1.Create,
							},
							Rule: admissionregistrationv1.Rule{
								APIGroups:   []string{"apps"},
								APIVersions: []string{"v1"},
								Resources:   []string{"deployments"},
							},
						},
					},
				},
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook-service",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app": "webhook",
				},
				Ports: []corev1.ServicePort{
					{
						Port: 443,
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "webhook-pod",
				Namespace: "default",
				Labels: map[string]string{
					"app": "webhook",
				},
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeMutatingWebhook(request)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "[]", result, "A normal MutatingWebhookConfiguration should not have any issues")
}

func TestAnalyzeMutatingWebhookWithNonExistentService(t *testing.T) {
	// Test a MutatingWebhookConfiguration with a non-existent service
	port := int32(443)
	clientset := fake.NewSimpleClientset(
		&admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-mutating-webhook",
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{
				{
					Name: "test-webhook",
					ClientConfig: admissionregistrationv1.WebhookClientConfig{
						Service: &admissionregistrationv1.ServiceReference{
							Namespace: "default",
							Name:      "non-existent-service",
							Port:      &port,
						},
						CABundle: []byte("test-ca-bundle"),
					},
					Rules: []admissionregistrationv1.RuleWithOperations{
						{
							Operations: []admissionregistrationv1.OperationType{
								admissionregistrationv1.Create,
							},
							Rule: admissionregistrationv1.Rule{
								APIGroups:   []string{"apps"},
								APIVersions: []string{"v1"},
								Resources:   []string{"deployments"},
							},
						},
					},
				},
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeMutatingWebhook(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "references service", "Should detect non-existent service")
	assert.Contains(t, result, "which does not exist", "Should detect non-existent service")
}
