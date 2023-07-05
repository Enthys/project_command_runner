package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	flag "github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/exp/maps"
	"gopkg.in/yaml.v3"
)

var (
	excludedProjects *[]string
	searchTags       *[]string
	excludeTags      *[]string
	configPath       *string
	command          *string
)

type Config struct {
	Projects map[string]Project `yaml:"projects"`
}

type Project struct {
	Name string   `yaml:"name,omitempty"`
	Path string   `yaml:"path"`
	Tags []string `yaml:"tags,omitempty"`
}

func init() {
	excludedProjects = flag.StringArrayP("exclude", "e", []string{}, "Projects in which the provided command should not be used in")
	configPath = flag.StringP("config", "c", "commander.yaml", "Path to the configuration file")
	command = flag.StringP("command", "X", "", "The command which has to be executed")
	searchTags = flag.StringArray("tag-search", []string{}, "Tags by which to find projects")
	excludeTags = flag.StringArray("tag-exclude", []string{}, "Tags by which to filter out projects")

	flag.Parse()

	if *command == "" {
		log.Fatalln("Please provide the '--command' argument.")
	}
}

func parseConfig(config *Config) error {
	file, err := os.OpenFile(*configPath, os.O_RDONLY, os.ModeAppend)
	if err != nil {
		return fmt.Errorf("failed to open file %s. Error: %w", *configPath, err)
	}

	b, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read configuration file. Error: %w", err)
	}

	if err = yaml.Unmarshal(b, config); err != nil {
		return fmt.Errorf("failed to parse configuration file. Error: %w", err)
	}

	return nil
}

func removeExcludedProjects(config *Config) {
	for _, excludedProject := range *excludedProjects {
		delete(config.Projects, excludedProject)
	}
}

func filterByTags(config *Config) {
	for _, searchTag := range *searchTags {
		for projectName, project := range config.Projects {
			remove := true
			for _, projectTag := range project.Tags {
				if projectTag == searchTag {
					remove = false
				}
			}

			if remove {
				delete(config.Projects, projectName)
			}
		}
	}
}

func excludeByTags(config *Config) {
	for projectName, project := range config.Projects {
		remove := false
		for _, projectTag := range project.Tags {
			for _, searchTag := range *excludeTags {
				if projectTag == searchTag {
					remove = true
				}
			}
		}

		if remove {
			delete(config.Projects, projectName)
		}
	}
}

func executeCommandInProjects(config *Config) {
	width, _, err := terminal.GetSize(0)
	separator := "------------------------------------------"
	if err == nil {
		separator = strings.Repeat("-", width)
	}

	projectErrors := map[string]error{}
	for projectName, project := range config.Projects {
		commandParts := []string{"/bin/sh", "-c", fmt.Sprintf("cd %s && %s", project.Path, *command)}

		color.Green(
			"Running command for project '%s'.\nCommand: %s\n",
			projectName,
			strings.Join(commandParts, " "),
		)

		commandCmd := exec.Command("/bin/sh", "-c", fmt.Sprintf("cd %s && %s", project.Path, *command))
		res, err := commandCmd.CombinedOutput()
		fmt.Printf("%s", string(res))
		if err != nil {
			projectErrors[projectName] = errors.New(string(res))
		}
		color.Blue("%s\n", separator)
	}

	if len(projectErrors) > 0 {
		color.Red("%s", separator)
	}

	for projectName, err := range projectErrors {
		fmt.Printf(
			"Encountered an error while running command in project '%s'.\nError:\n",
			projectName,
		)

		color.Red("%s", err.Error())
		color.Red("%s", separator)
	}

	if len(projectErrors) > 0 {
		k := maps.Keys(projectErrors)
		log.Fatalf("Projects [%s] encountered an error when running command.", strings.Join(k, ", "))
		os.Exit(1)
	}
}

func main() {
	var config Config
	if err := parseConfig(&config); err != nil {
		log.Fatalln(err)
	}

	if len(*searchTags) > 0 {
		filterByTags(&config)
	}

	excludeByTags(&config)
	removeExcludedProjects(&config)
	executeCommandInProjects(&config)
}
