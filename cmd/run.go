package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/dao"
	"github.com/alajmo/mani/core/print"
)

func runCmd(config *dao.Config, configErr *error) *cobra.Command {
	var dryRun bool
	var cwd bool
	var describe bool
	var allProjects bool
	var tags []string
	var projects []string
	var output string

	cmd := cobra.Command{
		Use:   "run <command> [flags]",
		Short: "Run commands",
		Long: `Run commands.

The commands are specified in a mani.yaml file along with the projects you can target.`,

		Example: `  # Run task 'pwd' for all projects
  mani run pwd --all-projects

  # Checkout branch 'development' for all projects that have tag 'backend'
  mani run checkout -t backend branch=development`,

		DisableFlagsInUseLine: true,
		Args:                  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			core.CheckIfError(*configErr)
			executeRun(args, config, output, describe, dryRun, cwd, allProjects, tags, projects)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) != 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}

			if *configErr != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			return config.GetCommands(), cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().BoolVar(&describe, "describe", true, "Print command information")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "don't execute any command, just print the output of the command to see what will be executed")
	cmd.Flags().BoolVarP(&cwd, "cwd", "k", false, "current working directory")
	cmd.Flags().BoolVarP(&allProjects, "all-projects", "a", false, "target all projects")
	cmd.Flags().StringSliceVarP(&tags, "tags", "t", []string{}, "target projects by their tag")
	cmd.Flags().StringSliceVarP(&projects, "projects", "p", []string{}, "target projects by their name")
	cmd.Flags().StringVarP(&output, "output", "o", "list", "Output list|table|markdown|html")

	err := cmd.RegisterFlagCompletionFunc("projects", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		projects := config.GetProjectNames()
		return projects, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("tags", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		tags := config.GetTags()
		return tags, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	err = cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if *configErr != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		valid := []string { "table", "markdown", "html" }
		return valid, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func executeRun(
	args []string,
	config *dao.Config,
	outputFlag string,
	describeFlag bool,
	dryRunFlag bool,
	cwdFlag bool,
	allProjectsFlag bool,
	tagsFlag []string,
	projectsFlag []string,
) {
	projects := config.FilterProjects(cwdFlag, allProjectsFlag, tagsFlag, projectsFlag)

	var commandNames []string
	var userArgs []string
	for _, arg := range args {
		if strings.Contains(arg, "=") {
			userArgs = append(userArgs, arg)
		} else {
			commandNames = append(commandNames, arg)
		}
	}

	for _, cmd := range commandNames {
		command, err := config.GetCommand(cmd)
		core.CheckIfError(err)

		runCommand(command, projects, userArgs, config, outputFlag, describeFlag, dryRunFlag)
	}
}

func runCommand(
	command *dao.Command,
	projects []dao.Project,
	userArgs []string,
	config *dao.Config,
	outputFlag string,
	describeFlag bool,
	dryRunFlag bool,
) {
	command.SetEnvList(userArgs, config.GetEnv())

	if describeFlag {
		print.PrintCommandBlocks([]dao.Command {*command})
	}

	spinner, err := dao.CommandSpinner()
	core.CheckIfError(err)

	err = spinner.Start()
	var outputs []dao.ProjectOutput
	for _, project := range projects {
		spinner.Message(fmt.Sprintf(" %v", project.Name))

		output, err := command.RunCmd(config.Path, config.Shell, project, command.EnvList, dryRunFlag)
		if err != nil {
			fmt.Println(err)
		}

		outputs = append(outputs, dao.ProjectOutput {
			ProjectName: project.Name,
			Output: output,
		})
	}

	err = spinner.Stop()
	core.CheckIfError(err)

	print.PrintRun(outputFlag, outputs)
}
