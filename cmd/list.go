package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists Pie-related Docker containers.",
	Long:  `Lists Docker containers whose names start with "pie-", showing their name, ID, and status.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("NAME\t\tID\t\tSTATUS")

		dockerCmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}\t{{.ID}}\t{{.Status}}")
		output, err := dockerCmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error listing Docker containers: %v\n", err)
			fmt.Println(string(output))
			return
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "pie-") {
				parts := strings.SplitN(line, "\t", 3)
				if len(parts) == 3 {
					name := strings.TrimPrefix(parts[0], "pie-")
					id := parts[1]
					status := parts[2]
					fmt.Printf("%s\t%s\t%s\n", name, id[:10], status) // Truncate ID to 10 chars
				} else {
					fmt.Println("Warning: Unexpected Docker output format for line:", line)
				}
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
