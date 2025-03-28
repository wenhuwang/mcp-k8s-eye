package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/wenhuwang/mcp-k8s-eye/pkg/mcp"
)

var (
	name    = "mcp-k8s-eye"
	version = "0.1.1"
)

var rootCmd = &cobra.Command{
	Use:   "mcp-k8s-eye [command] [options]",
	Short: "Use MCP to monitor your Kubernetes",
	Long: `
  Use MCP (Model Context Protocol) to monitor your Kubernetes 

  # show this help
  mcp-k8s-eye -h

  # shows version information
  mcp-k8s-eye --version

  # TODO: add more examples`,
	Run: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("version") {
			fmt.Println(version)
			return
		}
		mcpServer, err := mcp.NewServer(name, version)
		if err != nil {
			panic(err)
		}
		if err := mcpServer.ServeStdio(); err != nil && !errors.Is(err, context.Canceled) {
			panic(err)
		}
	},
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Print version information and quit")
	_ = viper.BindPFlags(rootCmd.Flags())
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}
