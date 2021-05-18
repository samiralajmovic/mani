package cmd

import (
	"fmt"
	color "github.com/logrusorgru/aurora"
	"github.com/alajmo/mani/core"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

func initCmd() *cobra.Command {
	var autoDiscovery bool

	cmd := cobra.Command{
		Use:   "init",
		Short: "Initialize a mani repository",
		Long: `Initialize a mani repository.

Creates a mani repository - a directory with configuration file mani.yaml and a .gitignore file.`,
		Example: `  # Basic example
  mani init

  # Skip auto-discovery of projects
  mani init --auto-discovery=false`,

		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runInit(args, autoDiscovery)
		},
	}

	cmd.Flags().BoolVar(&autoDiscovery, "auto-discovery", true, "walk current directory and find git repositories to add to mani.yaml")

	return &cmd
}

func runInit(args []string, autoDiscovery bool) {
	var configPath string
	if len(args) > 0 && filepath.IsAbs(args[0]) {
		configPath = args[0]
	} else if len(args) > 0 {
		wd, err := os.Getwd()
		core.CheckIfError(err)
		configPath = filepath.Join(wd, args[0])
	} else {
		wd, err := os.Getwd()
		core.CheckIfError(err)
		configPath = wd
	}

	os.MkdirAll(configPath, os.ModePerm)

	configFilepath := filepath.Join(configPath, "mani.yaml")
	if _, err := os.Stat(configFilepath); err == nil {
		fmt.Printf("fatal: %q is already a mani directory\n", configPath)
		os.Exit(1)
	}

	// Add to mani.yaml
	url := core.GetRemoteUrl(configPath)
	rootName := filepath.Base(configPath)
	rootPath := "."
	rootUrl := url
	rootProject := core.Project{Name: rootName, Path: rootPath, Url: rootUrl}
	projects := []core.Project{rootProject}
	if autoDiscovery {
		prs, err := core.FindVCSystems(configPath)

		if err != nil {
			fmt.Println(err)
		}

		projects = append(projects, prs...)
	}

	tmpl, err := template.New("default").Parse(`projects: {{ range .}}
  - name: {{ .Name }}
   {{ if ne .Path .Name }} path: {{ .Path }}{{ end }}
   {{- if .Url }} url: {{ .Url }} {{ end }}
{{ end }}
commands:
  - name: hello-world
    description: Print Hello World
    command: echo "Hello World"
`,
	)

	// Create mani.yaml
	f, err := os.Create(configFilepath)
	core.CheckIfError(err)

	err = tmpl.Execute(f, projects)
	core.CheckIfError(err)

	f.Close()
	fmt.Println(color.Green("\u2713"), "Initialized mani repository in", configPath)

	// Add gitignore file
	gitignoreFilepath := filepath.Join(configPath, ".gitignore")
	if _, err := os.Stat(gitignoreFilepath); os.IsNotExist(err) {
		err := ioutil.WriteFile(gitignoreFilepath, []byte(""), 0644)

		core.CheckIfError(err)
	}

	var projectNames []string
	for _, project := range projects {
		projectNames = append(projectNames, project.Name)
	}

	// Add projects to gitignore file
	err = core.UpdateProjectsToGitignore(projectNames, gitignoreFilepath)
	core.CheckIfError(err)
}
