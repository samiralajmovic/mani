package cmd

import (
	"github.com/alajmo/mani/core"
	"github.com/alajmo/mani/core/print"
	"github.com/spf13/cobra"
)

func listTagsCmd(configFile *string, listFlags *core.ListFlags) *cobra.Command {
	var tagFlags core.ListTagFlags
	var projects []string

	cmd := cobra.Command {
		Aliases: []string { "tag" },
		Use:   "tags [flags]",
		Short: "List tags",
		Long:  "List tags.",
		Example: `  # List tags
  mani list tags`,
		Run: func(cmd *cobra.Command, args []string) {
			listTags(configFile, args, listFlags, &tagFlags, projects)
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_, config, err := core.ReadConfig(*configFile)
			if err != nil {
				return []string{}, cobra.ShellCompDirectiveDefault
			}

			tags := core.GetTags(config.Projects)
			return tags, cobra.ShellCompDirectiveNoFileComp
		},
	}

	cmd.Flags().StringSliceVarP(&projects, "projects", "p", []string{}, "filter tags by their project")
	err := cmd.RegisterFlagCompletionFunc("projects", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, config, err := core.ReadConfig(*configFile)

		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		projects := core.GetProjectNames(config.Projects)
		return projects, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	cmd.Flags().StringSliceVar(&tagFlags.Headers, "headers", []string{ "name" }, "Specify headers, defaults to name, description")
	err = cmd.RegisterFlagCompletionFunc("headers", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		_, _, err := core.ReadConfig(*configFile)

		if err != nil {
			return []string{}, cobra.ShellCompDirectiveDefault
		}

		validHeaders := []string { "name" }

		return validHeaders, cobra.ShellCompDirectiveDefault
	})
	core.CheckIfError(err)

	return &cmd
}

func listTags(
	configFile *string,
	args []string,
	listFlags *core.ListFlags,
	tagFlags *core.ListTagFlags,
	projects []string,
) {
	_, config, err := core.ReadConfig(*configFile)
	core.CheckIfError(err)

	allTags := core.GetTags(config.Projects)
	if (len(args) == 0 && len(projects) == 0) {
		print.PrintTags(allTags, *listFlags, *tagFlags)
		return
	}

	if (len(args) > 0 && len(projects) == 0) {
		args = core.Intersection(args, allTags)
		print.PrintTags(args, *listFlags, *tagFlags)
	} else if (len(args) == 0 && len(projects) > 0) {
		projectTags := core.FilterTagOnProject(config.Projects, projects)
		print.PrintTags(projectTags, *listFlags, *tagFlags)
	} else {
		projectTags := core.FilterTagOnProject(config.Projects, projects)
		args = core.Intersection(args, projectTags)
		print.PrintTags(args, *listFlags, *tagFlags)
	}
}
