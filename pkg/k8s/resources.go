package k8s

import (
	"context"
	"fmt"
	"strings"
)

// GetResourceGV gets the Group and Version for a resource based on its Kind
// Returns:
// - group: API Group of the resource, empty string for core API
// - version: Version of the resource
// - error: Returns error if resource not found or other errors occur
func (k *Kubernetes) GetResourceGV(ctx context.Context, kind string) (group, version string, err error) {
	// Get all API resources
	apiResourceList, err := k.discoveryClient.ServerPreferredResources()
	if err != nil {
		return "", "", fmt.Errorf("failed to get API resources: %v", err)
	}

	// Iterate through all resource lists to find matching Kind
	for _, resources := range apiResourceList {
		// GroupVersion format is "group/version" or "version" (for core API)
		gv := strings.Split(resources.GroupVersion, "/")

		// Check each resource
		for _, resource := range resources.APIResources {
			if resource.Kind == kind {
				// Determine group and version based on GroupVersion format
				if len(gv) == 2 {
					// Non-core API: group/version
					return gv[0], gv[1], nil
				} else if len(gv) == 1 {
					// Core API: version only
					return "", gv[0], nil
				}
			}
		}
	}

	return "", "", fmt.Errorf("resource kind %q not found", kind)
}

// GetPreferredResourceGV gets the preferred Group and Version for a resource based on its Kind
// If a resource has multiple versions, returns the preferred version
func (k *Kubernetes) GetPreferredResourceGV(ctx context.Context, kind string) (group, version string, err error) {
	// Get all API groups
	apiGroups, _, err := k.discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return "", "", fmt.Errorf("failed to get API groups: %v", err)
	}

	// First try to find preferred version in API groups
	for _, apiGroup := range apiGroups {
		// Check preferred version for each group
		if len(apiGroup.PreferredVersion.Version) > 0 {
			// Use preferred version to check resources
			resourceList, err := k.discoveryClient.ServerResourcesForGroupVersion(
				fmt.Sprintf("%s/%s", apiGroup.Name, apiGroup.PreferredVersion.Version))
			if err != nil {
				continue
			}

			// Find matching Kind in resource list
			for _, resource := range resourceList.APIResources {
				if resource.Kind == kind {
					return apiGroup.Name, apiGroup.PreferredVersion.Version, nil
				}
			}
		}
	}

	// If not found in preferred versions, fall back to regular search
	return k.GetResourceGV(ctx, kind)
}
