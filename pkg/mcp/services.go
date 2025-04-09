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

func (s *Server) serviceAnalyze(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"].(string)
	res, err := s.k8s.AnalyzeServices(ctx, ns)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to analyze services in namespace %s: %v", ns, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}
