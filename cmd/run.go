package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <6-character-id>",
	Short: "Runs a Docker container by its 6-character ID.",
	Long:  `Runs a Docker container with a name in the format "pie-<6-character-id>".`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		containerID := args[0]
		if len(containerID) != 6 {
			fmt.Println("Error: The ID must be exactly 6 characters long.")
			return
		}

		containerName := "pie-" + containerID
		fmt.Printf("Attempting to run Pie container: %s\n", containerName)

		stopCmd := exec.Command("docker", "stop", containerName)
		stopCmd.Run()

		dockerCmd := exec.Command("docker", "start", "-a", containerName)
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr

		err := dockerCmd.Run()
		if err != nil {
			fmt.Printf("Error running container %s: %v\n", containerName, err)
		}

	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
