package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
)

func (s *Server) initIngress() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("ingress analyze",
				mcp.WithDescription("analyze ingress status"),
				mcp.WithString("namespace",
					mcp.Description("the ingress namespace to analyze"),
				),
			),
			Handler: s.ingressAnalyze,
		},
	}
}
func (s *Server) ingressAnalyze(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var ns string
	if v, ok := ctr.Params.Arguments["namespace"].(string); ok {
		ns = v
	}
	r := common.Request{
		Context:   ctx,
		Namespace: ns,
	}
	res, err := s.k8s.AnalyzeIngress(r)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to analyze ingress in namespace %s: %v", ns, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}
