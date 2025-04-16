package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
)

// newTestKubernetes creates a Kubernetes instance for testing with fake clientset
func newTestKubernetes(clientset kubernetes.Interface, dynamicClient dynamic.Interface) *Kubernetes {
	return &Kubernetes{
		clientset:                   clientset,
		dynamicClient:               dynamicClient,
		discoveryClient:             newTestDiscoveryClient(),
		deferredDiscoveryRESTMapper: restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(newTestDiscoveryClient())),
		config:                      &rest.Config{},
	}
}

// setupNormalPodsClientset creates a fake clientset with pod data
func newNormalPodsClientset() *fake.Clientset {
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
func newFailingPodsClientset() *fake.Clientset {
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

func TestAnalyzePods(t *testing.T) {
	t.Run("Analyze pods with failures", func(t *testing.T) {
		// Setup
		clientset := newFailingPodsClientset()
		k := newTestKubernetes(clientset, nil)

		// Execute
		result, err := k.AnalyzePod(context.Background(), "test-namespace")

		// Verify
		assert.NoError(t, err, "Should not return an error")
		assert.Contains(t, result, "test-namespace/pending-pod", "Result should contain the pending pod reference")
		assert.Contains(t, result, "test-namespace/crashloop-pod", "Result should contain the crashloop pod reference")
	})

	t.Run("Analyze pods with no failures", func(t *testing.T) {
		// Setup a healthy pod
		clientset := newNormalPodsClientset()
		k := newTestKubernetes(clientset, nil)

		// Execute
		result, err := k.AnalyzePod(context.Background(), "test-namespace")

		// Verify
		assert.NoError(t, err, "Should not return an error")
		assert.Equal(t, "[]", result, "Result should be an empty array for no failures")
	})
}

func TestAnalyzeContainerStatusFailures(t *testing.T) {
	t.Run("Analyze container in CrashLoopBackOff", func(t *testing.T) {
		// Setup
		clientset := fake.NewSimpleClientset()
		k := newTestKubernetes(clientset, nil)

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
		clientset := newNormalPodsClientset()
		k := newTestKubernetes(clientset, nil)

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
