package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
)

// Register networkpolicy analyze tool
func (s *Server) initNetworkPolicy() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("networkpolicy analyze",
				mcp.WithDescription("analyze networkpolicy status"),
				mcp.WithString("namespace",
					mcp.Description("the namespace to analyze network policies in"),
					mcp.Required(),
				),
			),
			Handler: s.networkPolicyAnalyze,
		},
	}
}

func (s *Server) networkPolicyAnalyze(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var ns string
	if v, ok := ctr.Params.Arguments["namespace"].(string); ok {
		ns = v
	}
	res, err := s.k8s.AnalyzeNetworkPolicy(common.Request{
		Context:   ctx,
		Namespace: ns,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to analyze network policies in namespace %s: %v", ns, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}
