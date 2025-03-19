package utils

import (
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// ResourceCleaner clean kubernetes resource default fields
type ResourceCleaner struct {
}

// NewResourceCleaner create a new ResourceCleaner instance
func NewResourceCleaner() *ResourceCleaner {
	return &ResourceCleaner{}
}

// Clean clean kubernetes resource default fields
func (rc *ResourceCleaner) Clean(obj runtime.Object) runtime.Object {
	if obj == nil {
		return nil
	}

	rc.cleanObjectMeta(obj)

	// clean specific fields based on resource type
	switch o := obj.(type) {
	case *corev1.Pod:
		rc.cleanPod(o)
	case *appsv1.Deployment:
		rc.cleanDeployment(o)
	case *appsv1.StatefulSet:
		rc.cleanStatefulSet(o)
	case *appsv1.DaemonSet:
		rc.cleanDaemonSet(o)
	case *corev1.Service:
		rc.cleanService(o)
	case *networkingv1.Ingress:
		rc.cleanIngress(o)
	case *corev1.ConfigMap:
		rc.cleanConfigMap(o)
	case *corev1.Secret:
		rc.cleanSecret(o)
	case *corev1.PersistentVolumeClaim:
		rc.cleanPVC(o)
	}

	return obj
}

// cleanObjectMeta clean object metadata
func (rc *ResourceCleaner) cleanObjectMeta(obj runtime.Object) {
	metaObj, ok := obj.(metav1.Object)
	if !ok {
		return
	}

	metaObj.SetAnnotations(nil)
	metaObj.SetOwnerReferences(nil)
	metaObj.SetManagedFields(nil)
	metaObj.SetResourceVersion("")
	metaObj.SetGeneration(0)
	metaObj.SetUID("")
}

// cleanPod clean Pod specific fields
func (rc *ResourceCleaner) cleanPod(pod *corev1.Pod) {
	pod.Status = corev1.PodStatus{}

	delete(pod.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
}

// cleanDeployment clean Deployment specific fields
func (rc *ResourceCleaner) cleanDeployment(deploy *appsv1.Deployment) {
	deploy.Status = appsv1.DeploymentStatus{}

	delete(deploy.Annotations, "deployment.kubernetes.io/revision")
	delete(deploy.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
}

// cleanStatefulSet clean StatefulSet specific fields
func (rc *ResourceCleaner) cleanStatefulSet(sts *appsv1.StatefulSet) {
	sts.Status = appsv1.StatefulSetStatus{}

	delete(sts.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
}

// cleanDaemonSet clean DaemonSet specific fields
func (rc *ResourceCleaner) cleanDaemonSet(ds *appsv1.DaemonSet) {
	ds.Status = appsv1.DaemonSetStatus{}

	delete(ds.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
}

// cleanService clean Service specific fields
func (rc *ResourceCleaner) cleanService(svc *corev1.Service) {
	svc.Status = corev1.ServiceStatus{}

	delete(svc.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
}

// cleanIngress clean Ingress specific fields
func (rc *ResourceCleaner) cleanIngress(ing *networkingv1.Ingress) {
	ing.Status = networkingv1.IngressStatus{}

	delete(ing.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
}

// cleanConfigMap clean ConfigMap specific fields
func (rc *ResourceCleaner) cleanConfigMap(cm *corev1.ConfigMap) {
	// ConfigMap has no special default fields to clean
	// just clean common metadata
	delete(cm.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
}

// cleanSecret clean Secret specific fields
func (rc *ResourceCleaner) cleanSecret(secret *corev1.Secret) {
	// Secret has no special default fields to clean
	// just clean common metadata
	delete(secret.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
}

// cleanPVC clean PersistentVolumeClaim specific fields
func (rc *ResourceCleaner) cleanPVC(pvc *corev1.PersistentVolumeClaim) {
	pvc.Status = corev1.PersistentVolumeClaimStatus{}

	delete(pvc.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
}

// CleanList clean default fields in resource list
func (rc *ResourceCleaner) CleanList(list runtime.Object) runtime.Object {
	if list == nil {
		return nil
	}

	listValue := reflect.ValueOf(list)
	if listValue.Kind() == reflect.Ptr {
		listValue = listValue.Elem()
	}

	// get Items field
	items := listValue.FieldByName("Items")
	if !items.IsValid() {
		return list
	}

	// iterate and clean each item
	for i := 0; i < items.Len(); i++ {
		item := items.Index(i).Addr().Interface().(runtime.Object)
		rc.Clean(item)
	}

	return list
}
