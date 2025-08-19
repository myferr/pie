package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

var requirements string

var rootCmd = &cobra.Command{
	Use:   "pie [file]",
	Short: "A Python dependency manager and runner using Docker.",
	Long:  `Pie is a CLI tool that simplifies Python project setup by managing dependencies and running scripts in isolated Docker environments.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config := readPieConfig()

		// Determine the file to run
		fileToRun := config.Main
		if len(args) > 0 && args[0] != "." {
			fileToRun = args[0]
		}

		config.Main = fileToRun // Update config for Piefile generation

		// Handle requirements flag
		if requirements != "" {
			config.DepsFile = requirements
			config.Dependencies = nil // Ensure deps_file takes precedence
		} else if config.DepsFile == "" && len(config.Dependencies) == 0 {
			fmt.Println("No dependencies specified. Scanning for imports...")
			scanDependencies()
			config.DepsFile = "PieDeps.txt"
		}

		generatePiefile(config)
		runDocker(config)
	},
}

func scanDependencies() {
	var deps []string
	importRegex := regexp.MustCompile(`^(?:from|import)\s+([\w\.]+)`)

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".py") {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			matches := importRegex.FindAllStringSubmatch(string(content), -1)
			for _, match := range matches {
				deps = append(deps, match[1])
			}
		}
		return nil
	})

	// Create PieDeps.txt
	pieDepsContent := strings.Join(deps, "\n")
	ioutil.WriteFile("PieDeps.txt", []byte(pieDepsContent), 0644)
	fmt.Println("Created PieDeps.txt")
}

func init() {
	rootCmd.Flags().StringVarP(&requirements, "requirements", "r", "", "Specify a custom requirements file")
}

func runDocker(config PieConfig) {
	fmt.Println("Building Docker image...")
	buildCmd := exec.Command("docker", "build", "-t", "pie-runner", "-f", ".pie/Piefile", ".")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		fmt.Println("Error building Docker image:", err)
		return
	}

	fmt.Println("Running Python script in Docker container...")
	runCmd := exec.Command("docker", "run", "pie-runner")
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	if err := runCmd.Run(); err != nil {
		fmt.Println("Error running Docker container:", err)
		return
	}
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
