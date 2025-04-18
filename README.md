## mcp-k8s-eye

mcp-k8s-eye is a tool that can manage kubernetes cluster and analyze workload status.

## Features

- [x] Connect to a Kubernetes cluster
- [x] **Generic Kubernetes Resources** management capabilities
  - Support all navtie resources: Pod, Deployment, Service, StatefulSet, Ingress...
  - Support CustomResourceDefinition resources
  - Operations include: list, get, create, update, delete
- [x] Pod management capabilities (exec, logs)
- [x] Deployment management capabilities (scale)
- [x] Analyze pods
- [x] Analyze services
- [x] Analyze deployments
- [x] Analyze statefulsets
- [x] Analyze ingress
- [x] Analyze nodes
- [ ] Analyze cluster


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
### STDIO
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

### SSE
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