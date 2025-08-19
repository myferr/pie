package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new Pie project.",
	Long:  `Creates a pie.yml file and a .pie/Piefile to configure the Python environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing Pie project...")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
