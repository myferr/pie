package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

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
}

func init() {
	rootCmd.Flags().StringVarP(&requirements, "requirements", "r", "", "Specify a custom requirements file")
}

func runDocker(config PieConfig) {
	buildCmd := exec.Command("docker", "build", "-t", "pie-runner", "-f", ".pie/Piefile", ".")
	if err := buildCmd.Run(); err != nil {
		fmt.Println("Error building Docker image:", err)
		return
	}

	runCmd := exec.Command("docker", "run", "--name", "pie-runner-container", "pie-runner")
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-
c
		fmt.Println("\nStopping container...")
		stopCmd := exec.Command("docker", "stop", "pie-runner-container")
		stopCmd.Run()
		rmCmd := exec.Command("docker", "rm", "pie-runner-container")
		rmCmd.Run()
		fmt.Println("Container stopped and removed.")
		os.Exit(0)
	}()

	if err := runCmd.Run(); err != nil {
		// Don't print error on graceful shutdown
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				if status.Signaled() && status.Signal() == syscall.SIGTERM {
					return
				}
			}
		}
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
