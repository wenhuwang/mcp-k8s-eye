package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (s *Server) initResource() []server.ServerTool {
	return []server.ServerTool{
		{
			Tool: mcp.NewTool("resource list",
				mcp.WithDescription("list resources in a namespace or all namespaces"),
				mcp.WithString("kind",
					mcp.Description("the kind of resource to list"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("the namespace to list resources in"),
					mcp.Required(),
				),
			),
			Handler: s.resourceList,
		},
		{
			Tool: mcp.NewTool("resource get",
				mcp.WithDescription("get resource details"),
				mcp.WithString("kind",
					mcp.Description("the kind of resource to get"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("the namespace to get resources in"),
					mcp.Required(),
				),
				mcp.WithString("name",
					mcp.Description("the resource name to get"),
					mcp.Required(),
				),
			),
			Handler: s.resourceGet,
		},
		{
			Tool: mcp.NewTool("resource delete",
				mcp.WithDescription("delete resource"),
				mcp.WithString("kind",
					mcp.Description("the kind of resource to delete"),
					mcp.Required(),
				),
				mcp.WithString("namespace",
					mcp.Description("the namespace to get resources in"),
					mcp.Required(),
				),
				mcp.WithString("name",
					mcp.Description("the resource name to delete"),
					mcp.Required(),
				),
			),
			Handler: s.resourceDelete,
		},
	}
}

func (s *Server) resourceList(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"].(string)
	kind := ctr.Params.Arguments["kind"].(string)
	res, err := s.k8s.ResourceList(ctx, kind, ns)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list resources in namespace %s: %v", ns, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

func (s *Server) resourceGet(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"].(string)
	kind := ctr.Params.Arguments["kind"].(string)
	name := ctr.Params.Arguments["name"].(string)
	res, err := s.k8s.ResourceGet(ctx, kind, ns, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get resource %s/%s: %v", ns, name, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

func (s *Server) resourceDelete(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ns := ctr.Params.Arguments["namespace"].(string)
	kind := ctr.Params.Arguments["kind"].(string)
	name := ctr.Params.Arguments["name"].(string)
	res, err := s.k8s.ResourceDelete(ctx, kind, ns, name)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to delete resource %s/%s: %v", ns, name, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}
