package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initNode() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("node analyze",
				mcp.WithDescription("analyze node status"),
				mcp.WithString("name",
					mcp.Description("the node name to analyze"),
				),
			),
			Handler: s.nodeAnalyze,
		},
	}
}
func (s *Server) nodeAnalyze(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var name string
	if v, ok := ctr.Params.Arguments["name"].(string); ok {
		name = v
	}
	res, err := s.k8s.AnalyzeNode(ctx, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to analyze node %s: %v", name, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}
