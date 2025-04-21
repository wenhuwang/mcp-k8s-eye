package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
)

// Register cronjob analyze tool
func (s *Server) initCronJob() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("cronjob analyze",
				mcp.WithDescription("analyze cronjob status"),
				mcp.WithString("namespace",
					mcp.Description("the cronjob namespace to analyze"),
					mcp.Required(),
				),
			),
			Handler: s.cronjobAnalyze,
		},
	}
}

func (s *Server) cronjobAnalyze(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var ns string
	if v, ok := ctr.Params.Arguments["namespace"].(string); ok {
		ns = v
	}
	res, err := s.k8s.AnalyzeCronJob(common.Request{
		Context:   ctx,
		Namespace: ns,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to analyze cronjob in namespace %s: %v", ns, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}
