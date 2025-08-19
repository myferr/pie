package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "pie",
	Short: "A Python dependency manager and runner using Docker.",
	Long:  `Pie is a CLI tool that simplifies Python project setup by managing dependencies and running scripts in isolated Docker environments.`,
	Run: func(cmd *cobra.Command, args []string) {
		// This is the main execution logic for `pie <file>` or `pie .`
		fmt.Println("Pie executed with args: ", args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
