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

var ingressClassName = "ingressclass1"

func TestAnalyzeIngress(t *testing.T) {

	clientset := fake.NewSimpleClientset(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress1",
				Namespace: "default",
			},
			Spec: networkingv1.IngressSpec{
				IngressClassName: &ingressClassName,
			},
		},

		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress2",
				Namespace: "default",
				Annotations: map[string]string{
					"kubernetes.io/ingress.class": "ingressclass2",
				},
			},
		},
		&networkingv1.IngressClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ingressclass1",
			},
		},
		&networkingv1.IngressClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ingressclass2",
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeIngress(request)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "[]", result)
}

func TestAnalyzeIngressWithoutIngressClass(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress1",
				Namespace: "default",
			},
		},
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress2",
				Namespace: "default",
				Annotations: map[string]string{
					"kubernetes.io/ingress.class": "ingressclass2",
				},
			},
		},
	)
	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeIngress(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "Ingress does not specify an ingress class")
	assert.Contains(t, result, "Ingress uses the ingress class ingressclass2 which does not exist")
}

func TestAnalyzeIngressWithoutService(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress1",
				Namespace: "default",
			},
			Spec: networkingv1.IngressSpec{
				IngressClassName: &ingressClassName,
				Rules: []networkingv1.IngressRule{
					networkingv1.IngressRule{
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Path: "/",
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: "service1",
											},
										},
									},
									{
										Path: "/aaa",
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: "service2",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		&networkingv1.IngressClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ingressclass1",
			},
		},
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "service2",
				Namespace: "default",
			},
		},
	)

	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}

	result, err := k.AnalyzeIngress(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "Ingress uses the service default/service1 which does not exist")
}

func TestAnalyzeIngressWithoutSecret(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ingress1",
				Namespace: "default",
			},
			Spec: networkingv1.IngressSpec{
				IngressClassName: &ingressClassName,
				TLS: []networkingv1.IngressTLS{
					{
						SecretName: "secret1",
					},
					{
						SecretName: "secret2",
					},
				},
			},
		},
		&networkingv1.IngressClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "ingressclass1",
			},
		},
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "secret2",
				Namespace: "default",
			},
		},
	)
	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}
	result, err := k.AnalyzeIngress(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "Ingress uses the secret default/secret1 which does not exist")
}
