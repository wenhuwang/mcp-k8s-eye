package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/common"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/mcp"
)

var rootCmd = &cobra.Command{
	Use:   "mcp-k8s-eye [command] [options]",
	Short: "Use MCP to manage and analyze your Kubernetes",
	Long: `
  Use MCP (Model Context Protocol) to monitor your Kubernetes 

  # show this help
  mcp-k8s-eye -h

  # shows version information
  mcp-k8s-eye --version

  # start STDIO server
  mcp-k8s-eye

  # Start SSE server
  mcp-k8s-eye --sse
}

  # TODO: add more examples`,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("version") {
			fmt.Println(common.Version)
			return
		}
		mcpServer, err := mcp.NewServer(common.ProjectName, common.Version)
		if err != nil {
			log.Fatalf("Failed to create MCP server: %v", err)
		}

		if viper.GetBool("sse") {
			sse := mcpServer.ServeSSE()
			defer sse.Shutdown(cmd.Context())

			var port int
			if port = viper.GetInt("port"); port == 0 {
				port = common.DefaultSSEPort
			}
			log.Printf("Starting SSE server on port: %d", port)
			if err := sse.Start(fmt.Sprintf(":%d", port)); err != nil {
				log.Fatalf("Failed to start SSE server: %v", err)
				return
			}
			return
		}

		if err := mcpServer.ServeStdio(); err != nil && !errors.Is(err, context.Canceled) {
			log.Fatalf("Failed to start STDIO server: %v", err)
		}
	},
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Print version information and quit")
	rootCmd.Flags().BoolP("sse", "s", false, "Start SSE server(default port is 8080)")
	rootCmd.Flags().IntP("port", "p", 0, "Start SSE server with specified port")
	_ = viper.BindPFlags(rootCmd.Flags())
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
