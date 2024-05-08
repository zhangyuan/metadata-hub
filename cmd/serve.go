package cmd

import (
	"log"
	"metadata-hub/pkg/serve"

	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use: "serve",
	Run: func(cmd *cobra.Command, args []string) {
		if err := serve.Invoke(configDirectory, storeDirectory, addr); err != nil {
			log.Fatalln(err)
		}
	},
}

var configDirectory string
var addr string
var storeDirectory string

func init() {
	rootCmd.AddCommand(serveCmd)

	serveCmd.Flags().StringVarP(&configDirectory, "config-directory", "c", "", "path to config directory")
	_ = serveCmd.MarkFlagRequired("config-directory")

	serveCmd.Flags().StringVarP(&storeDirectory, "store-directory", "s", "", "path to store directory")

	serveCmd.Flags().StringVarP(&addr, "listening addr", "l", ":8080", "listening addr")
}
