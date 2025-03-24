package k8s

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	testing_k8s "k8s.io/client-go/testing"
)

// newTestKubernetes creates a Kubernetes instance for testing with fake clientset
func newTestKubernetes(clientset kubernetes.Interface) *Kubernetes {
	return &Kubernetes{
		clientset: clientset,
		config:    &rest.Config{},
	}
}

// setupMockClientset creates a fake clientset with pod data
func setupMockClientset() *fake.Clientset {
	// Create a mock pod to use in tests
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:  "test-container",
					Ready: true,
					State: v1.ContainerState{
						Running: &v1.ContainerStateRunning{},
					},
				},
			},
		},
	}

	// Create a fake clientset with the pod
	clientset := fake.NewSimpleClientset(pod)
	return clientset
}

// setupFailingPods creates a fake clientset with failing pods
func setupFailingPods() *fake.Clientset {
	// Create pods with different failure scenarios
	pendingPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pending-pod",
			Namespace: "test-namespace",
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				{
					Type:    v1.PodScheduled,
					Status:  v1.ConditionFalse,
					Reason:  "Unschedulable",
					Message: "0/3 nodes are available: 3 Insufficient memory",
				},
			},
		},
	}

	crashLoopPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crashloop-pod",
			Namespace: "test-namespace",
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:  "crash-container",
					Ready: false,
					State: v1.ContainerState{
						Waiting: &v1.ContainerStateWaiting{
							Reason:  "CrashLoopBackOff",
							Message: "Back-off 5m0s restarting failed container",
						},
					},
					LastTerminationState: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason: "OOMKilled",
						},
					},
				},
			},
		},
	}

	// Create a fake clientset with the pods
	clientset := fake.NewSimpleClientset(pendingPod, crashLoopPod)
	return clientset
}

func TestPodList(t *testing.T) {
	t.Run("List pods successfully", func(t *testing.T) {
		// Setup
		clientset := setupMockClientset()
		k := newTestKubernetes(clientset)

		// Execute
		result, err := k.PodList(context.Background(), "test-namespace")

		// Verify
		assert.NoError(t, err, "Should not return an error")
		assert.Contains(t, result, "test-pod", "Result should contain the pod name")
		assert.Contains(t, result, "test-namespace", "Result should contain the namespace")
	})

	t.Run("Handle error when listing pods", func(t *testing.T) {
		// Setup a clientset that will return an error
		clientset := fake.NewSimpleClientset()
		// Inject an error using a reactor
		clientset.PrependReactor("list", "pods", func(action testing_k8s.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, errors.New("failed to list pods")
		})

		k := newTestKubernetes(clientset)

		// Execute
		result, err := k.PodList(context.Background(), "test-namespace")

		// Verify
		assert.Error(t, err, "Should return an error")
		assert.Equal(t, "", result, "Result should be empty")
		assert.Contains(t, err.Error(), "failed to list pods", "Error message should contain the original error")
	})
}

func TestPodGet(t *testing.T) {
	t.Run("Get pod successfully", func(t *testing.T) {
		// Setup
		clientset := setupMockClientset()
		k := newTestKubernetes(clientset)

		// Execute
		result, err := k.PodGet(context.Background(), "test-namespace", "test-pod")

		// Verify
		assert.NoError(t, err, "Should not return an error")
		assert.Contains(t, result, "test-pod", "Result should contain the pod name")
		assert.Contains(t, result, "test-namespace", "Result should contain the namespace")
	})

	t.Run("Handle error when getting non-existent pod", func(t *testing.T) {
		// Setup
		clientset := setupMockClientset()
		k := newTestKubernetes(clientset)

		// Execute
		result, err := k.PodGet(context.Background(), "test-namespace", "non-existent-pod")

		// Verify
		assert.Error(t, err, "Should return an error")
		assert.Equal(t, "", result, "Result should be empty")
		assert.Contains(t, err.Error(), "not found", "Error message should indicate the pod wasn't found")
	})
}

func TestPodDelete(t *testing.T) {
	t.Run("Delete pod successfully", func(t *testing.T) {
		// Setup
		clientset := setupMockClientset()
		k := newTestKubernetes(clientset)

		// Execute
		result, err := k.PodDelete(context.Background(), "test-namespace", "test-pod")

		// Verify
		assert.NoError(t, err, "Should not return an error")
		assert.Equal(t, "Pod deleted successfully", result, "Result should indicate success")

		// Verify the pod was actually deleted
		_, err = clientset.CoreV1().Pods("test-namespace").Get(context.Background(), "test-pod", metav1.GetOptions{})
		assert.Error(t, err, "Pod should be deleted")
		assert.Contains(t, err.Error(), "not found", "Error should indicate the pod wasn't found")
	})

	t.Run("Handle error when deleting non-existent pod", func(t *testing.T) {
		// Setup
		clientset := setupMockClientset()
		k := newTestKubernetes(clientset)

		// Execute
		result, err := k.PodDelete(context.Background(), "test-namespace", "non-existent-pod")

		// Verify
		assert.Error(t, err, "Should return an error")
		assert.Equal(t, "", result, "Result should be empty")
		assert.Contains(t, err.Error(), "not found", "Error message should indicate the pod wasn't found")
	})
}

func TestAnalyzePods(t *testing.T) {
	t.Run("Analyze pods with failures", func(t *testing.T) {
		// Setup
		clientset := setupFailingPods()
		k := newTestKubernetes(clientset)

		// Execute
		result, err := k.AnalyzePods(context.Background(), "test-namespace")

		// Verify
		assert.NoError(t, err, "Should not return an error")
		assert.Contains(t, result, "test-namespace/pending-pod", "Result should contain the pending pod reference")
		assert.Contains(t, result, "test-namespace/crashloop-pod", "Result should contain the crashloop pod reference")
	})

	t.Run("Analyze pods with no failures", func(t *testing.T) {
		// Setup a healthy pod
		clientset := setupMockClientset()
		k := newTestKubernetes(clientset)

		// Execute
		result, err := k.AnalyzePods(context.Background(), "test-namespace")

		// Verify
		assert.NoError(t, err, "Should not return an error")
		assert.Equal(t, "[]", result, "Result should be an empty array for no failures")
	})

	t.Run("Handle error when listing pods", func(t *testing.T) {
		// Setup a clientset that will return an error
		clientset := fake.NewSimpleClientset()
		// Inject an error using a reactor
		clientset.PrependReactor("list", "pods", func(action testing_k8s.Action) (handled bool, ret runtime.Object, err error) {
			return true, nil, errors.New("failed to list pods")
		})

		k := newTestKubernetes(clientset)

		// Execute
		result, err := k.AnalyzePods(context.Background(), "test-namespace")

		// Verify
		assert.Error(t, err, "Should return an error")
		assert.Equal(t, "", result, "Result should be empty")
		assert.Contains(t, err.Error(), "failed to list pods", "Error message should contain the original error")
	})
}

func TestAnalyzeContainerStatusFailures(t *testing.T) {
	t.Run("Analyze container in CrashLoopBackOff", func(t *testing.T) {
		// Setup
		clientset := setupMockClientset()
		k := newTestKubernetes(clientset)

		// Create container statuses with a crash loop
		containerStatuses := []v1.ContainerStatus{
			{
				Name:  "crash-container",
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "CrashLoopBackOff",
						Message: "Back-off 5m0s restarting failed container",
					},
				},
				LastTerminationState: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						Reason: "OOMKilled",
					},
				},
			},
		}

		// Execute
		failures := k.analyzeContainerStatusFailures(containerStatuses, "test-pod", "test-namespace", "Running")

		// Verify
		assert.NotEmpty(t, failures, "Should detect failures")
		assert.Contains(t, failures[0].Text, "OOMKilled", "Should identify OOMKilled as termination reason")
		assert.Contains(t, failures[0].Text, "crash-container", "Should mention the container name")
	})

	t.Run("Analyze container with waiting error", func(t *testing.T) {
		// Setup
		clientset := setupMockClientset()
		k := newTestKubernetes(clientset)

		// Create container statuses with a waiting error
		containerStatuses := []v1.ContainerStatus{
			{
				Name:  "error-container",
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "ImagePullBackOff",
						Message: "Back-off pulling image error",
					},
				},
			},
		}

		// Execute
		failures := k.analyzeContainerStatusFailures(containerStatuses, "test-pod", "test-namespace", "Pending")

		// Verify
		assert.NotEmpty(t, failures, "Should detect failures")
		assert.Equal(t, "Back-off pulling image error", failures[0].Text, "Should contain the error message")
	})
}
