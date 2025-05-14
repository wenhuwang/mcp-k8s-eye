## mcp-k8s-eye

mcp-k8s-eye is a tool that can manage kubernetes cluster and analyze workload status.

## Features

### Core Kubernetes Operations
- [x] Connect to a Kubernetes cluster
- [x] **Generic Kubernetes Resources** management capabilities
  - Support all navtie resources: Pod, Deployment, Service, StatefulSet, Ingress...
  - Support CustomResourceDefinition resources
  - Operations include: list, get, create, update, delete
- [x] Pod management capabilities (exec, logs)
- [x] Deployment management capabilities (scale)
- [ ] Describe Kubernetes resources
- [ ] Explain Kubernetes resources


### Diagnostics
- [x] Pod diagnostics (analyze pod status, container status, pod resource utilization)
- [x] Service diagnostics (analyze service selector configuration, not ready endpoints, events)
- [x] Deployment diagnostics (analyze available replicas)
- [x] StatefulSet diagnostics (analyze statefulset service if exists, pvc if exists, available replicas)
- [x] CronJob diagnostics (analyze cronjob schedule, starting deadline, last schedule time)
- [x] Ingress diagnostics (analyze ingress class configuration, related services, tls secrets)
- [x] NetworkPolicy diagnostics (analyze networkpolicy configuration, affected pods)
- [x] ValidatingWebhook diagnostics (analyze webhook configuration, referenced services and pods)
- [x] MutatingWebhook diagnostics (analyze webhook configuration, referenced services and pods)
- [x] Node diagnostics (analyze node conditions)
- [ ] Cluster diagnostics and troubleshooting 

### Monitoring
- [x] Pod, Deployment, ReplicaSet, StatefulSet, DaemonSet workload resource usage (cpu, memory)
- [ ] Node capacity, utilization (cpu, memory)
- [ ] Cluster capacity, utilization (cpu, memory)

### Advanced Features
- [x] Multiple transport protocols support (Stdio, SSE)
- [x] Support multiple AI Clients


## Requirements
- Go 1.23 or higher
- kubectl configured

## Installation
```
# clone the repository
git clone https://github.com/wenhuwang/mcp-k8s-eye.git
cd mcp-k8s-eye

# build the binary
go build -o mcp-k8s-eye
```

## Usage
### Stdio mode
```
{
  "mcpServers": {
    "k8s eye": {
      "command": "YOUR mcp-k8s-eye PATH",
      "env": {
        "HOME": "USER HOME DIR"
      },
    }
  }
}
```
`env.HOME` is used to set the HOME directory for kubeconfig file.

### SSE mode
1. start your mcp sse server
2. config your mcp server

```
{
  "mcpServers": {
    "k8s eye": {
      "url": "http://localhost:8080/sse",
      "env": {}
    }
  }
}
```

### cursor tools
![cursor tools](./images/mcp-server-tools.png)