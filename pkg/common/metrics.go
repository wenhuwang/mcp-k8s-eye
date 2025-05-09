package common

type PodFormattedMetrics struct {
	Name       string
	Namespace  string
	Containers []ContainerFormattedMetrics
}

type ContainerFormattedMetrics struct {
	Name        string
	CPUUsage    string
	MemoryUsage string
}
