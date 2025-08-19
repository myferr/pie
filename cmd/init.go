package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes a new Pie project.",
	Long:  `Creates a pie.yml file to configure the Python environment.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Initializing Pie project...")

		// Create pie.yml
		pieYmlContent := `main: main.py
python: "3.12"
# You can specify a dependency file, or list your dependencies here
# deps_file: PieDeps.yaml
dependencies:
  - pandas
  - numpy
`
		err := ioutil.WriteFile("pie.yml", []byte(pieYmlContent), 0644)
		if err != nil {
			fmt.Println("Error creating pie.yml:", err)
			return
		}
		fmt.Println("Created pie.yml")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
