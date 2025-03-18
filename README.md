## mcp-k8s-eye

mcp-k8s-eye is a tool that can manage kubernetes cluster and analyze workload status.

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
- [x] List pods by namespace
- [x] Get pod
- [x] Get pod logs
- [x] Delete pod
- [x] Exec command in pod
- [x] Analyze pods
- [ ] List all services
- [ ] List all deployments
- [ ] List all nodes
- [ ] Analyze services
- [ ] Analyze deployments
- [ ] Analyze ingress
- [ ] Analyze nodes
- [ ] Analyze cluster