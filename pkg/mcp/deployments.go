package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initDeployments() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("deployment scale",
				mcp.WithDescription("scale deployment replicas"),
				mcp.WithString("namespace",
					mcp.Description("the namespace of the deployment"),
				),
				mcp.WithString("deployment",
					mcp.Description("the deployment to scale"),
				),
				mcp.WithNumber("replicas",
					mcp.Description("the number of replicas to scale to"),
				),
			),
			Handler: s.deploymentScale,
		},
		{
			Tool: mcp.NewTool("deployment analyze",
				mcp.WithDescription("analyze deployment status"),
				mcp.WithString("namespace",
					mcp.Description("the namespace to analyze deployments in"),
				),
			),
			Handler: s.deploymentAnalyze,
		},
	}
}

func (s *Server) deploymentScale(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"].(string)
	deploy := ctr.Params.Arguments["deployment"].(string)
	replicas := int32(ctr.Params.Arguments["replicas"].(float64))
	res, err := s.k8s.DeploymentScale(ctx, ns, deploy, replicas)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to scale deployment %s/%s: %v", ns, deploy, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

func (s *Server) deploymentAnalyze(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"].(string)
	res, err := s.k8s.AnalyzeDeployments(ctx, ns)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to analyze deployments in namespace %s: %v", ns, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}
