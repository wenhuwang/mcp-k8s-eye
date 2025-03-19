## mcp-k8s-eye

mcp-k8s-eye is a tool that can manage kubernetes cluster and analyze workload status.

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
```
{
  "mcpServers": {
    "kubernetes": {
      "command": "YOUR mcp-k8s-eye PATH",
      "env": {
        "HOME": "YOUR HOME DIR"
      },
    }
  }
}
```

## Features

- [x] Connect to a Kubernetes cluster
- [x] Pod management capabilities (list, get, exec, logs, delete)
- [x] Deployment management capabilities (list, get, scale, delete)
- [x] Service management capabilities (list, get, delete)
- [ ] StatefulSet management capabilities (list, get, delete)
- [ ] DaemonSet management capabilities (list, get, delete)
- [ ] Ingress management capabilities (list, get, delete)
- [ ] Node management capabilities (list, get, delete)
- [x] Analyze pods
- [x] Analyze services
- [x] Analyze deployments
- [ ] Analyze statefulsets
- [ ] Analyze daemonsets
- [ ] Analyze ingress
- [ ] Analyze nodes
- [ ] Analyze cluster