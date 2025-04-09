package k8s

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	discoveryfake "k8s.io/client-go/discovery/fake"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubernetesfake "k8s.io/client-go/kubernetes/fake"
)

type resourceMapEntry struct {
	list *metav1.APIResourceList
	err  error
}

type fakeDiscovery struct {
	*discoveryfake.FakeDiscovery

	lock         sync.Mutex
	groupList    *metav1.APIGroupList
	groupListErr error
	resourceMap  map[string]*resourceMapEntry
}

func (c *fakeDiscovery) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if rl, ok := c.resourceMap[groupVersion]; ok {
		return rl.list, rl.err
	}
	return nil, errors.New("doesn't exist")
}

func (c *fakeDiscovery) ServerGroups() (*metav1.APIGroupList, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.groupList == nil {
		return nil, errors.New("doesn't exist")
	}
	return c.groupList, c.groupListErr
}
func newTestDiscoveryClient() *fakeDiscovery {
	fake := &fakeDiscovery{
		groupList: &metav1.APIGroupList{
			Groups: []metav1.APIGroup{
				{
					Name: "apps",
					Versions: []metav1.GroupVersionForDiscovery{
						{
							GroupVersion: "apps/v1",
							Version:      "v1",
						},
					},
				},
			},
		},
		resourceMap: map[string]*resourceMapEntry{
			"apps/v1": {
				list: &metav1.APIResourceList{
					GroupVersion: "apps/v1",
					APIResources: []metav1.APIResource{{
						Name:         "deployments",
						SingularName: "deployment",
						Namespaced:   true,
						Kind:         "Deployment",
						ShortNames:   []string{"deploy"},
					}},
				},
			},
		},
	}
	return fake
}

func newDynamicClient() *dynamicfake.FakeDynamicClient {
	unstructuredDeployment := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"namespace": "test-namespace",
				"name":      "test-deployment",
			},
		},
	}
	scheme := runtime.NewScheme()
	dynamicClient := dynamicfake.NewSimpleDynamicClient(scheme, unstructuredDeployment)
	return dynamicClient
}

func TestResourceList(t *testing.T) {
	t.Run("List resources successfully", func(t *testing.T) {
		// Setup
		clientset := kubernetesfake.NewSimpleClientset(&appsv1.Deployment{})
		k := newTestKubernetes(clientset, newDynamicClient())

		// Execute
		result, err := k.ResourceList(context.Background(), "Deployment", "test-namespace")

		// Verify
		assert.NoError(t, err, "Should not return an error")
		assert.Contains(t, result, "test-deployment", "Result should contain the deployment name")
		assert.Contains(t, result, "test-namespace", "Result should contain the namespace")
	})
}

func TestResourceGet(t *testing.T) {
	t.Run("Get resource successfully", func(t *testing.T) {
		// Setup
		clientset := kubernetesfake.NewSimpleClientset(&appsv1.Deployment{})
		k := newTestKubernetes(clientset, newDynamicClient())

		// Execute
		result, err := k.ResourceGet(context.Background(), "Deployment", "test-namespace", "test-deployment")

		// Verify
		assert.NoError(t, err, "Should not return an error")
		assert.Contains(t, result, "test-deployment", "Result should contain the deployment name")
		assert.Contains(t, result, "test-namespace", "Result should contain the namespace")
	})

	t.Run("Handle error when getting non-existent deployment", func(t *testing.T) {
		// Setup
		clientset := kubernetesfake.NewSimpleClientset(&appsv1.Deployment{})
		k := newTestKubernetes(clientset, newDynamicClient())

		// Execute
		result, err := k.ResourceGet(context.Background(), "Deployment", "test-namespace", "non-existent-deployment")

		// Verify
		assert.Error(t, err, "Should return an error")
		assert.Equal(t, "", result, "Result should be empty")
		assert.Contains(t, err.Error(), "not found", "Error message should indicate the deployment wasn't found")
	})
}

func TestResourceDelete(t *testing.T) {
	t.Run("Delete resource successfully", func(t *testing.T) {
		// Setup
		clientset := kubernetesfake.NewSimpleClientset(&appsv1.Deployment{})
		k := newTestKubernetes(clientset, newDynamicClient())

		// Execute
		result, err := k.ResourceDelete(context.Background(), "Deployment", "test-namespace", "test-deployment")

		// Verify
		assert.NoError(t, err, "Should not return an error")
		assert.Contains(t, result, "deleted successfully", "Result should indicate success")

		// Verify the pod was actually deleted
		_, err = k.ResourceGet(context.Background(), "Deployment", "test-namespace", "test-deployment")
		assert.Error(t, err, "Deployment should be deleted")
		assert.Contains(t, err.Error(), "not found", "Error should indicate the deployment wasn't found")
	})

	t.Run("Handle error when deleting non-existent deployment", func(t *testing.T) {
		// Setup
		clientset := kubernetesfake.NewSimpleClientset(&appsv1.Deployment{})
		k := newTestKubernetes(clientset, newDynamicClient())

		// Execute
		result, err := k.ResourceDelete(context.Background(), "Deployment", "test-namespace", "non-existent-deployment")

		// Verify
		assert.Error(t, err, "Should return an error")
		assert.Equal(t, "", result, "Result should be empty")
		assert.Contains(t, err.Error(), "not found", "Error message should indicate the deployment wasn't found")
	})
}
