package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
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
		{
			Tool: mcp.NewTool("resource create or update",
				mcp.WithDescription("create or update resource"),
				mcp.WithString("resource",
					mcp.Description("the resource to create or update"),
					mcp.Required(),
				),
			),
			Handler: s.resourceCreateOrUpdate,
		},
		{
			Tool: mcp.NewTool("workload resource usage",
				mcp.WithDescription("workload resource usage"),
				mcp.WithString("kind",
					mcp.Description("the kind of workload"),
					mcp.Required(),
					mcp.Enum("Deployment", "StatefulSet", "DaemonSet", "ReplicaSet", "Pod"),
				),
				mcp.WithString("namespace",
					mcp.Description("the namespace of workload"),
					mcp.Required(),
				),
				mcp.WithString("name",
					mcp.Description("the name of workload"),
				),
			),
			Handler: s.workloadResourceUsage,
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

func (s *Server) resourceCreateOrUpdate(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	resource := ctr.Params.Arguments["resource"].(string)
	res, err := s.k8s.ResourceCreateOrUpdate(ctx, resource)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to create/update resource: %v", err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

func (s *Server) workloadResourceUsage(ctx context.Context, ctr mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	namespace := ctr.Params.Arguments["namespace"].(string)
	kind := ctr.Params.Arguments["kind"].(string)
	var name string
	if v, ok := ctr.Params.Arguments["name"].(string); ok {
		name = v
	}
	res, err := s.k8s.WorkloadResourceUsage(common.Request{
		Context:      ctx,
		Name:         name,
		Namespace:    namespace,
		WorkloadType: kind,
	})
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to get resource usage in namesoace %s: %v", namespace, err)), nil
	}
	return mcp.NewToolResultText(res), nil
}

// test prompt
func (s *Server) getNamespacePrompt(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	var name string
	if v, ok := request.Params.Arguments["name"]; ok {
		name = v
	} else {
		name = "all namespaces"
	}

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Get namespace %s", name),
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Text: fmt.Sprintf("Get namespace %s", name),
				},
			},
			{
				Role: mcp.RoleAssistant,
				Content: mcp.TextContent{
					Text: fmt.Sprintf("Namespace %s:", name),
				},
			},
		},
	}, nil
}
