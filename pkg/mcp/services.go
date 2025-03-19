package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initServices() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("service list",
				mcp.WithDescription("list services in a namespace"),
				mcp.WithString("namespace",
					mcp.Description("the namespace to list services in"),
				),
			),
			Handler: s.serviceList,
		},
		{
			Tool: mcp.NewTool("service get",
				mcp.WithDescription("get service details"),
				mcp.WithString("namespace",
					mcp.Description("the namespace to get service from"),
				),
				mcp.WithString("service",
					mcp.Description("the service to get"),
				),
			),
			Handler: s.serviceGet,
		},
		{
			Tool: mcp.NewTool("service delete",
				mcp.WithDescription("delete service"),
				mcp.WithString("namespace",
					mcp.Description("the namespace to delete service from"),
				),
				mcp.WithString("service",
					mcp.Description("the service to delete"),
				),
			),
			Handler: s.serviceDelete,
		},
		{
			Tool: mcp.NewTool("service analyze",
				mcp.WithDescription("analyze service status"),
				mcp.WithString("namespace",
					mcp.Description("the namespace to analyze services in"),
				),
			),
			Handler: s.serviceAnalyze,
		},
	}
}

func (s *Server) serviceList(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"].(string)
	res, err := s.k8s.ServiceList(ctx, ns)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list services in namespace %s: %v", ns, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

func (s *Server) serviceGet(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"].(string)
	svc := ctr.Params.Arguments["service"].(string)
	res, err := s.k8s.ServiceGet(ctx, ns, svc)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get service %s/%s: %v", ns, svc, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

func (s *Server) serviceDelete(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"].(string)
	svc := ctr.Params.Arguments["service"].(string)
	res, err := s.k8s.ServiceDelete(ctx, ns, svc)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete service %s/%s: %v", ns, svc, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

func (s *Server) serviceAnalyze(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"].(string)
	res, err := s.k8s.AnalyzeServices(ctx, ns)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to analyze services in namespace %s: %v", ns, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}
