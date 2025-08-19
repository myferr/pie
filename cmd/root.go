package cmd

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type PieConfig struct {
	Main         string   `yaml:"main"`
	Python       string   `yaml:"python"`
	DepsFile     string   `yaml:"deps_file,omitempty"`
	Dependencies []string `yaml:"dependencies,omitempty"`
}

var file string
var pythonVersion string
var dependencies string
var verbose bool
var dispose bool

var rootCmd = &cobra.Command{
	Use:   "pie [file]",
	Short: "A Python dependency manager and runner using Docker.",
	Long:  `Pie is a CLI tool that simplifies Python project setup by managing dependencies and running scripts in isolated Docker environments.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		config := PieConfig{}

		// Prioritize flags over pie.yml
		if pythonVersion != "" {
			config.Python = pythonVersion
		} else {
			config.Python = "latest"
		}

		if dependencies != "" {
			// Trim braces and split
			deps := strings.Trim(dependencies, "{}")
			config.Dependencies = strings.Split(deps, ",")
		}

		// If no flags, read pie.yml
		if pythonVersion == "" && dependencies == "" {
			config = readPieConfig()
		}

		// Determine the file to run
		fileToRun := config.Main
		if len(args) > 0 && args[0] != "." {
			fileToRun = args[0]
		}

		config.Main = fileToRun // Update config for Piefile generation

		generatePiefile(config)
		runDocker(config)
	},
}

func init() {
	rootCmd.Flags().StringVarP(&file, "file", "f", "pie.yml", "Specify a custom pie.yml file")
	rootCmd.Flags().StringVarP(&pythonVersion, "python-version", "p", "", "Specify the Python version")
	rootCmd.Flags().StringVarP(&dependencies, "dependencies", "d", "", "Specify a comma-separated list of dependencies")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full Docker build output")
	rootCmd.Flags().BoolVar(&dispose, "dispose", false, "Remove the container after it runs")
}

func runDocker(config PieConfig) {
	buildCmd := exec.Command("docker", "build", "-t", "pie-runner", "-f", ".pie/Piefile", ".")
	if verbose {
		buildCmd.Stdout = os.Stdout
		buildCmd.Stderr = os.Stderr
	}
	if err := buildCmd.Run(); err != nil {
		fmt.Println("Error building Docker image:", err)
		return
	}

	rand.Seed(time.Now().UnixNano())
	containerName := fmt.Sprintf("pie-%s", randomString(6))

	args := []string{"run", "--name", containerName}
	if dispose {
		args = append(args, "--rm")
	}
	args = append(args, "pie-runner")

	runCmd := exec.Command("docker", args...)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-
c
		fmt.Println("\nStopping container...")
		stopCmd := exec.Command("docker", "stop", containerName)
		stopCmd.Run()
		rmCmd := exec.Command("docker", "rm", containerName)
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

	if _, err := os.Stat(file); err == nil {
		data, err := ioutil.ReadFile(file)
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

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}