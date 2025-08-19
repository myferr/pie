package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type PieConfig struct {
	Main         string   `yaml:"main"`
	Python       string   `yaml:"python"`
	DepsFile     string   `yaml:"deps_file,omitempty"`
	Dependencies []string `yaml:"dependencies,omitempty"`
}

var rootCmd = &cobra.Command{
	Use:   "pie",
	Short: "A Python dependency manager and runner using Docker.",
	Long:  `Pie is a CLI tool that simplifies Python project setup by managing dependencies and running scripts in isolated Docker environments.`,
	Run: func(cmd *cobra.Command, args []string) {
		config := readPieConfig()
		generatePiefile(config)
	},
}

func generatePiefile(config PieConfig) {
	// Create .pie directory if it doesn't exist
	if _, err := os.Stat(".pie"); os.IsNotExist(err) {
		os.Mkdir(".pie", 0755)
	}

	// Generate Dockerfile content
	var dockerfileContent string
	dockerfileContent += fmt.Sprintf("FROM python:%s-slim\n", config.Python)
	dockerfileContent += "WORKDIR /app\n"
	dockerfileContent += "COPY . .\n"

	if len(config.Dependencies) > 0 {
		dockerfileContent += fmt.Sprintf("RUN pip install %s\n", strings.Join(config.Dependencies, " "))
	} else if config.DepsFile != "" {
		dockerfileContent += fmt.Sprintf("RUN pip install -r %s\n", config.DepsFile)
	}

	dockerfileContent += fmt.Sprintf("CMD [\"python\", \"%s\"]\n", config.Main)

	// Write content to .pie/Piefile
	err := ioutil.WriteFile(".pie/Piefile", []byte(dockerfileContent), 0644)
	if err != nil {
		fmt.Println("Error creating .pie/Piefile:", err)
		return
	}
	fmt.Println("Generated .pie/Piefile")
}

func readPieConfig() PieConfig {
	config := PieConfig{
		Python: "latest", // Default python version
	}

	if _, err := os.Stat("pie.yml"); err == nil {
		data, err := ioutil.ReadFile("pie.yml")
		if err != nil {
			fmt.Println("Error reading pie.yml:", err)
			os.Exit(1)
		}

		err = yaml.Unmarshal(data, &config)
		if err != nil {
			fmt.Println("Error parsing pie.yml:", err)
			os.Exit(1)
		}
	}

	return config
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
