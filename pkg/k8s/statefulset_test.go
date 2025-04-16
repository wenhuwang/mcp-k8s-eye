package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestAnalyzeStatefulSet(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		})
	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}
	result, err := k.AnalyzeStatefulSet(request)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "[]", result)
}

func TestAnalyzeStatefulSetWithoutService(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				ServiceName: "test-service",
			},
		},
	)
	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}
	result, err := k.AnalyzeStatefulSet(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "StatefulSet uses the service default/test-service which does not exist")
}

func TestAnalyzeStatefulSetWithoutPVC(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test-pvc",
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
	result, err := k.AnalyzeStatefulSet(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "StatefulSet uses the pvc default/test-pvc which does not exist")
}

func TestAnalyzeStatefulSetNamespaceFiltering(t *testing.T) {
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1",
				Namespace: "default",
			},
		},
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test2",
				Namespace: "other",
			},
			Spec: appsv1.StatefulSetSpec{
				ServiceName: "test-service",
			},
		},
	)
	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}
	result, err := k.AnalyzeStatefulSet(request)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "[]", result)
}

func TestAnalyzeStatefulSetReplicas(t *testing.T) {
	replicas := int32(3)
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &replicas,
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-0",
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-1",
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-2",
				Namespace: "default",
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
	result, err := k.AnalyzeStatefulSet(request)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, "[]", result)
}

func TestAnalyzeStatefulSetUnavailableReplicas(t *testing.T) {
	replicas := int32(3)
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &replicas,
			},
			Status: appsv1.StatefulSetStatus{
				AvailableReplicas: 0,
			},
		})
	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}
	result, err := k.AnalyzeStatefulSet(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "StatefulSet has 3 replicas, but only 0 pods are running")
}

func TestAnalyzeStatefulSetUnavailableReplicasWithPodPending(t *testing.T) {
	replicas := int32(3)
	clientset := fake.NewSimpleClientset(
		&appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: &replicas,
			},
			Status: appsv1.StatefulSetStatus{
				AvailableReplicas: 1,
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-0",
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-1",
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodPending,
			},
		},
	)
	k := newTestKubernetes(clientset, nil)
	request := common.Request{
		Context:   context.Background(),
		Namespace: "default",
	}
	result, err := k.AnalyzeStatefulSet(request)
	if err != nil {
		t.Error(err)
	}
	assert.Contains(t, result, "StatefulSet pod default/test-1 is not in Running state")
}
