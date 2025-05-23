package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ovh/cds/cli"
	"github.com/ovh/cds/sdk"
	"github.com/ovh/cds/sdk/cdsclient"
)

var projectCmd = cli.Command{
	Name:    "project",
	Aliases: []string{"projects"},
	Short:   "Manage CDS project",
}

func projectCommands() []*cobra.Command {
	return []*cobra.Command{
		cli.NewListCommand(projectListCmd, projectListRun, nil, withAllCommandModifiers()...),
		cli.NewGetCommand(projectShowCmd, projectShowRun, nil, withAllCommandModifiers()...),
		cli.NewCommand(projectCreateCmd, projectCreateRun, nil),
		cli.NewDeleteCommand(projectDeleteCmd, projectDeleteRun, nil, withAllCommandModifiers()...),
		projectKey(),
		projectVariable(),
		projectVCS(),
		projectIntegration(),
	}
}

func project() *cobra.Command { return cli.NewCommand(projectCmd, nil, projectCommands()) }

func projectShell() *cobra.Command {
	return cli.NewCommand(projectCmd, nil, append(projectCommands(),
		application(),
		workflow(),
		environment(),
	))
}

var projectListCmd = cli.Command{
	Name:  "list",
	Short: "List CDS projects",
}

func projectListRun(v cli.Values) (cli.ListResult, error) {
	projs, err := client.ProjectList(false, false)
	if err != nil {
		return nil, err
	}
	return cli.AsListResult(projs), nil
}

var projectShowCmd = cli.Command{
	Name:  "show",
	Short: "Show a CDS project",
	Ctx: []cli.Arg{
		{Name: _ProjectKey},
	},
}

func projectShowRun(v cli.Values) (interface{}, error) {
	mods := []cdsclient.RequestModifier{}
	mods = append(mods, func(r *http.Request) {
		q := r.URL.Query()
		q.Set("withWorkflowNames", "true")
		q.Set("withIntegrations", "true")
		r.URL.RawQuery = q.Encode()
	})

	proj, err := client.ProjectGet(v.GetString(_ProjectKey), mods...)
	if err != nil {
		return nil, err
	}

	var p = struct {
		Key          string `cli:"key,key"`
		Name         string `cli:"name"`
		Description  string `cli:"description"`
		URL          string `cli:"url"`
		API          string `cli:"api"`
		Workflows    string `cli:"workflows"`
		NbWorkflows  int    `cli:"nb_workflows"`
		RepoManagers string `cli:"repositories_manager"`
		Integrations string `cli:"integration"`
	}{
		Key:         proj.Key,
		Name:        proj.Name,
		Description: proj.Description,
		NbWorkflows: len(proj.WorkflowNames),
		Workflows:   cli.Ellipsis(strings.Join(proj.WorkflowNames.Names(), ","), 70),
		URL:         proj.URLs.UIURL,
		API:         proj.URLs.APIURL,
	}

	var integrations []string
	for _, inte := range proj.Integrations {
		integrations = append(integrations, inte.Name)
	}
	p.Integrations = cli.Ellipsis(strings.Join(integrations, ","), 70)

	var repomanagers []string
	for _, vcs := range proj.VCSServers {
		repomanagers = append(repomanagers, vcs.Name)
	}
	p.RepoManagers = cli.Ellipsis(strings.Join(repomanagers, ","), 70)

	return p, nil
}

var projectCreateCmd = cli.Command{
	Name:  "create",
	Short: "Create a CDS project",
	Ctx: []cli.Arg{
		{Name: _ProjectKey},
		{Name: "project-name"},
	},
	OptionalArgs: []cli.Arg{
		{Name: "group-name"},
	},
	Aliases: []string{"add"},
}

func projectCreateRun(v cli.Values) error {
	proj := &sdk.Project{
		Name: v.GetString("project-name"),
		Key:  v.GetString(_ProjectKey),
	}

	groupName := v.GetString("group-name")
	if groupName != "" {
		proj.ProjectGroups = []sdk.GroupPermission{
			{
				Group: sdk.Group{Name: groupName},
			},
		}
	}

	return client.ProjectCreate(proj)
}

var projectDeleteCmd = cli.Command{
	Name:  "delete",
	Short: "Delete a CDS project",
	Ctx: []cli.Arg{
		{Name: _ProjectKey},
	},
}

func projectDeleteRun(v cli.Values) error {
	projKey := v.GetString(_ProjectKey)
	if v.GetBool("force") {
		// Delete all workflow
		ws, err := client.WorkflowList(projKey)
		if err != nil && !sdk.ErrorIs(err, sdk.ErrNoProject) {
			return err
		}
		for _, w := range ws {
			if err := client.WorkflowDelete(projKey, w.Name); err != nil && !sdk.ErrorIs(err, sdk.ErrNoProject) {
				return err
			}
		}

		// Delete all apps
		apps, err := client.ApplicationList(projKey)
		if err != nil && !sdk.ErrorIs(err, sdk.ErrNoProject) {
			return err
		}
		for _, app := range apps {
			if err := client.ApplicationDelete(projKey, app.Name); err != nil && !sdk.ErrorIs(err, sdk.ErrNoProject) {
				return err
			}
		}

		// Delete all pipelines
		pips, err := client.PipelineList(projKey)
		if err != nil && !sdk.ErrorIs(err, sdk.ErrNoProject) {
			return err
		}
		for _, pip := range pips {
			if err := client.PipelineDelete(projKey, pip.Name); err != nil && !sdk.ErrorIs(err, sdk.ErrNoProject) {
				return err
			}
		}

		// Delete all environments
		envs, err := client.EnvironmentList(projKey)
		if err != nil && !sdk.ErrorIs(err, sdk.ErrNoProject) {
			return err
		}
		for _, env := range envs {
			if err := client.EnvironmentDelete(projKey, env.Name); err != nil && !sdk.ErrorIs(err, sdk.ErrNoProject) {
				return err
			}
		}
	}

	if err := client.ProjectDelete(projKey); err != nil {
		if v.GetBool("force") && sdk.ErrorIs(err, sdk.ErrNoProject) {
			fmt.Println(err.Error())
			os.Exit(0)
		}
		return err
	}
	return nil
}
