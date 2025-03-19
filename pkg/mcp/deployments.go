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
			Tool: mcp.NewTool("deployment list",
				mcp.WithDescription("list deployments in a namespace"),
				mcp.WithString("namespace",
					mcp.Description("the namespace to list deployments in"),
				),
			),
			Handler: s.deploymentList,
		},
		{
			Tool: mcp.NewTool("deployment get",
				mcp.WithDescription("get deployment details"),
				mcp.WithString("namespace",
					mcp.Description("the namespace to get deployment from"),
				),
				mcp.WithString("deployment",
					mcp.Description("the deployment to get"),
				),
			),
			Handler: s.deploymentGet,
		},
		{
			Tool: mcp.NewTool("deployment delete",
				mcp.WithDescription("delete deployment"),
				mcp.WithString("namespace",
					mcp.Description("the namespace to delete deployment from"),
				),
				mcp.WithString("deployment",
					mcp.Description("the deployment to delete"),
				),
			),
			Handler: s.deploymentDelete,
		},
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

func (s *Server) deploymentList(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"].(string)
	res, err := s.k8s.DeploymentList(ctx, ns)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list deployments in namespace %s: %v", ns, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

func (s *Server) deploymentGet(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"].(string)
	deploy := ctr.Params.Arguments["deployment"].(string)
	res, err := s.k8s.DeploymentGet(ctx, ns, deploy)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get deployment %s/%s: %v", ns, deploy, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

func (s *Server) deploymentDelete(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"].(string)
	deploy := ctr.Params.Arguments["deployment"].(string)
	res, err := s.k8s.DeploymentDelete(ctx, ns, deploy)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete deployment %s/%s: %v", ns, deploy, err)), nil
	}
	return mcp.NewToolResultText(res), nil
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
