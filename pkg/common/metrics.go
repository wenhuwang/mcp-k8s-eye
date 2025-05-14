package common

// type PodFormattedMetrics struct {
// 	Name       string                      `json:"name"`
// 	Namespace  string                      `json:"namespace"`
// 	Containers []ContainerFormattedMetrics `json:"containers"`
// }

// type ContainerFormattedMetrics struct {
// 	Name        string `json:"name"`
// 	CPUUsage    string `json:"cpu"`
// 	MemoryUsage string `json:"memory"`
// }

type WorkloadFormattedMetrics struct {
	Name        string `json:"name"`
	Namespace   string `json:"namespace"`
	CPUUsage    string `json:"cpu"`
	MemoryUsage string `json:"memory"`
}
