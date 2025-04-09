package utils

import (
	"fmt"
	"testing"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		v1, v2 string
		want   bool
	}{
		{"v2", "v1", true},
		{"v1", "v2", false},
		{"v1alpha1", "v1beta1", false},
		{"v1beta1", "v1alpha1", true},
		{"v1alpha1", "v1alpha2", false},
		{"v1alpha2", "v1alpha1", true},
		{"v1beta1", "v2alpha1", false},
		{"v2alpha1", "v1beta1", true},
		{"v1", "v2alpha1", false},
		{"v2alpha1", "v1", true},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s vs %s", tc.v1, tc.v2), func(t *testing.T) {
			result := compareVersions(tc.v1, tc.v2)
			if result != tc.want {
				t.Errorf("compareVersions(%s, %s) = %t, want %t", tc.v1, tc.v2, result, tc.want)
			}
		})
	}
}

func TestGetGroupVersionForKind(t *testing.T) {
	cases := []struct {
		kind string
		want schema.GroupVersion
	}{
		{"Pod", schema.GroupVersion{Group: "", Version: "v1"}},
		{"Deployment", schema.GroupVersion{Group: "apps", Version: "v1"}},
		{"ReplicaSet", schema.GroupVersion{Group: "apps", Version: "v1"}},
		{"StatefulSet", schema.GroupVersion{Group: "apps", Version: "v1"}},
		{"DaemonSet", schema.GroupVersion{Group: "apps", Version: "v1"}},
		{"Job", schema.GroupVersion{Group: "batch", Version: "v1"}},
		{"CronJob", schema.GroupVersion{Group: "batch", Version: "v1"}},
		{"Service", schema.GroupVersion{Group: "", Version: "v1"}},
		{"Ingress", schema.GroupVersion{Group: "networking.k8s.io", Version: "v1"}},
		{"Endpoints", schema.GroupVersion{Group: "", Version: "v1"}},
		{"ConfigMap", schema.GroupVersion{Group: "", Version: "v1"}},
		{"Secret", schema.GroupVersion{Group: "", Version: "v1"}},
		{"PersistentVolume", schema.GroupVersion{Group: "", Version: "v1"}},
		{"PersistentVolumeClaim", schema.GroupVersion{Group: "", Version: "v1"}},
		{"Node", schema.GroupVersion{Group: "", Version: "v1"}},
		{"Namespace", schema.GroupVersion{Group: "", Version: "v1"}},
		{"Role", schema.GroupVersion{Group: "rbac.authorization.k8s.io", Version: "v1"}},
		{"RoleBinding", schema.GroupVersion{Group: "rbac.authorization.k8s.io", Version: "v1"}},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprintf("GetGroupVersionForKind(%s)", tc.kind), func(t *testing.T) {
			result := GetGroupVersionForKind(tc.kind)
			// t.Cleanup(func() {
			// 	t.Logf("GetGroupVersionForKind(%s) = %v, want %v", tc.kind, result, tc.want)
			// })
			if result != tc.want {
				t.Errorf("GetGroupVersionForKind(%s) = %v, want %v", tc.kind, result, tc.want)
			}
		})
	}
}
