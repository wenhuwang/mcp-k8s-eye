package utils

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
)

// compareVersions compares two Kubernetes API versions
// Returns true if v1 is newer than v2
func compareVersions(v1, v2 string) bool {
	// Extract the main version number before any alpha/beta suffix
	v1Main := strings.Split(strings.TrimPrefix(v1, "v"), "alpha")[0]
	v1Main = strings.Split(v1Main, "beta")[0]

	v2Main := strings.Split(strings.TrimPrefix(v2, "v"), "alpha")[0]
	v2Main = strings.Split(v2Main, "beta")[0]

	// Compare major version first
	if v1Main != v2Main {
		return v1Main > v2Main
	}

	// If we get here, the main version numbers are the same
	// Now compare alpha/beta status
	v1Weight := getVersionWeight(v1)
	v2Weight := getVersionWeight(v2)

	if v1Weight != v2Weight {
		return v1Weight > v2Weight
	}

	// If both are alpha or both are beta, compare their numbers
	if v1Weight < 2 { // alpha or beta
		v1Num := getPreReleaseNum(v1)
		v2Num := getPreReleaseNum(v2)
		return v1Num > v2Num
	}

	return false
}

// getVersionWeight returns a weight for alpha/beta/stable versions
// stable > beta > alpha
func getVersionWeight(version string) int {
	if strings.Contains(version, "alpha") {
		return 0
	}
	if strings.Contains(version, "beta") {
		return 1
	}
	return 2
}

// getPreReleaseNum extracts the number after alpha/beta
// For example: v1beta2 returns 2, v1alpha1 returns 1
func getPreReleaseNum(version string) int {
	if strings.Contains(version, "alpha") {
		parts := strings.Split(version, "alpha")
		if len(parts) > 1 {
			return parseVersionPart(parts[1])
		}
	}
	if strings.Contains(version, "beta") {
		parts := strings.Split(version, "beta")
		if len(parts) > 1 {
			return parseVersionPart(parts[1])
		}
	}
	return 0
}

// parseVersionPart safely parses a version string part to int
func parseVersionPart(part string) int {
	num := 0
	fmt.Sscanf(part, "%d", &num)
	return num
}

type GroupPriority struct {
	Group    string
	Priority int
}

var groupPriorities = map[string]int{
	"apps":              20,
	"networking.k8s.io": 15,
	"extensions":        10,
}

// GetGroupVersionForKind returns the GroupVersion for a given Kind using client-go's scheme
// Note: This only works for built-in resources registered in the default scheme
func GetGroupVersionForKind(kind string) schema.GroupVersion {
	s := clientsetscheme.Scheme

	// Get all known types from the scheme
	knownTypes := s.AllKnownTypes()

	var latestGV schema.GroupVersion
	var latestVersion string

	// Iterate through all known types to find matching kind
	for gvk := range knownTypes {
		if gvk.Kind == kind {
			// For the first match, initialize the latest version
			if latestVersion == "" {
				latestGV = gvk.GroupVersion()
				latestVersion = gvk.Version
				continue
			}

			// Compare versions if in the same group
			if gvk.Group == latestGV.Group {
				// Use proper version comparison
				if compareVersions(gvk.Version, latestVersion) {
					latestGV = gvk.GroupVersion()
					latestVersion = gvk.Version
				}
			} else {
				// Compare group priorities
				groupPriority := groupPriorities[gvk.Group]
				latestGroupPriority := groupPriorities[latestGV.Group]
				if groupPriority > latestGroupPriority {
					latestGV = gvk.GroupVersion()
					latestVersion = gvk.Version
				}
			}
		}

	}

	return latestGV
}
