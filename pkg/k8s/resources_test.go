package k8s

import (
	"context"
	"testing"
)

func TestGetResourceGV(t *testing.T) {
	k, err := NewKubernetes()
	if err != nil {
		t.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	testCases := []struct {
		kind        string
		wantGroup   string
		wantVersion string
	}{
		// {"Pod", "", "v1"},
		// {"Deployment", "apps", "v1"},
		// {"DaemonSet", "apps", "v1"},
		// {"StatefulSet", "apps", "v1"},
		// {"Job", "batch", "v1"},
		// {"CronJob", "batch", "v1"},
		// {"Service", "", "v1"},
		// {"ConfigMap", "", "v1"},
		// {"Secret", "", "v1"},
		// {"Ingress", "networking.k8s.io", "v1"},
		{"PersistentVolume", "storage.k8s.io", "v1"},
		{"PersistentVolumeClaim", "storage.k8s.io", "v1"},
		{"Node", "", "v1"},
		{"Namespace", "", "v1"},
		{"Event", "", "v1"},
		{"Endpoint", "discovery.k8s.io", "v1"},
	}

	for _, tc := range testCases {
		t.Run(tc.kind, func(t *testing.T) {
			group, version, err := k.GetResourceGV(context.Background(), tc.kind)
			if err != nil {
				t.Fatalf("Failed to get resource GV: %v", err)
			}

			if group != tc.wantGroup {
				t.Errorf("Group: got %s, want %s", group, tc.wantGroup)
			}

			if version != tc.wantVersion {
				t.Errorf("Version: got %s, want %s", version, tc.wantVersion)
			}
		})
	}
}
