package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/bulaya-ute/appctl/internal/db"
	"github.com/bulaya-ute/appctl/internal/deploy"
)

func newAppsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apps",
		Short: "Manage registered applications",
	}
	cmd.AddCommand(
		newAppsListCmd(),
		newAppsAddCmd(),
		newAppsShowCmd(),
		newAppsUpdateCmd(),
		newAppsRemoveCmd(),
		newAppsUnitCmd(),
	)
	return cmd
}

func newAppsListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all registered apps",
		RunE: func(cmd *cobra.Command, args []string) error {
			var apps []db.App
			if err := apiGet("/apps", &apps); err != nil {
				return err
			}
			if len(apps) == 0 {
				fmt.Println("No apps registered. Use 'appctl apps add' to register one.")
				return nil
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "NAME\tTYPE\tSOURCE\tDOMAIN\tSERVICE")
			for _, a := range apps {
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
					a.Name, a.Type, a.Source, a.Domain, a.ServiceName)
			}
			return tw.Flush()
		},
	}
}

func newAppsAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Register a new app (interactive)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("\n  Register new app")
			fmt.Println("  ────────────────")

			a := db.App{}
			a.Name = promptRequired("Name (slug, e.g. openplan-api)")
			a.Description = prompt("Description", "")
			a.Type = db.AppType(promptChoice("Type", []string{
				"dotnet-api", "react-spa", "static", "custom",
			}))
			a.Source = db.SourceType(promptChoice("Source", []string{"git", "local"}))

			if a.Source == db.SourceGit {
				a.GitRepoURL = promptRequired("Git repo URL")
				a.GitTokenPath = prompt("Path to file containing git token (leave blank for public repos)", "")
				a.Branch = prompt("Branch", "main")
			}

			a.LocalPath = promptRequired("Local path (where to clone / where files live)")

			if a.Type == db.AppTypeDotnetAPI || a.Type == db.AppTypeCustom {
				a.ServiceName = prompt("Systemd service name (leave blank if not a service)", "")
				if a.BindingPort, _ = strconv.Atoi(prompt("Binding port (leave blank if none)", "")); a.BindingPort < 0 {
					a.BindingPort = 0
				}
			}

			a.Domain = prompt("Domain for Caddy (leave blank to skip)", "")
			a.BuildCommand = prompt("Build command override (leave blank to use default for type)", "")
			a.RunCommand = prompt("Run command override for systemd ExecStart (leave blank to auto-detect)", "")
			a.PublishDir = prompt("Publish/dist dir override (leave blank to use default)", "")
			a.WebhookSecret = prompt("GitHub webhook secret (leave blank to skip webhook auth)", "")

			var created db.App
			if err := apiPost("/apps", a, &created); err != nil {
				return err
			}
			fmt.Printf("\n  ✓  Registered %q (id: %s)\n\n", created.Name, created.ID)

			if created.ServiceName != "" && promptYN("Create systemd unit file now?") {
				// We need DB access for this — use the daemon's deploy path instead.
				// For now, call the deploy endpoint which handles unit creation on first deploy.
				fmt.Println("  Tip: run 'appctl apps unit <name>' to write the systemd unit file.")
			}

			if promptYN("Deploy now?") {
				var dep db.Deployment
				if err := apiPost("/apps/"+created.Name+"/deploy",
					map[string]string{"version": "latest"}, &dep); err != nil {
					return err
				}
				fmt.Printf("  ✓  Deploy started (id: %s). Run 'appctl logs %s' to follow.\n\n",
					dep.ID, created.Name)
			}
			return nil
		},
	}
}

func newAppsShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show details for an app",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var a db.App
			if err := apiGet("/apps/"+args[0], &a); err != nil {
				return err
			}
			data, _ := json.MarshalIndent(a, "", "  ")
			fmt.Println(string(data))
			return nil
		},
	}
}

func newAppsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Update app configuration fields",
		Args:  cobra.ExactArgs(1),
	}
	flags := cmd.Flags()
	var (
		desc        string
		branch      string
		domain      string
		port        int
		service     string
		buildCmd    string
		runCmd      string
		publishDir  string
		gitRepo     string
		gitToken    string
		webhookSec  string
	)
	flags.StringVar(&desc, "description", "", "")
	flags.StringVar(&branch, "branch", "", "")
	flags.StringVar(&domain, "domain", "", "")
	flags.IntVar(&port, "port", 0, "")
	flags.StringVar(&service, "service", "", "systemd service name")
	flags.StringVar(&buildCmd, "build-command", "", "")
	flags.StringVar(&runCmd, "run-command", "", "")
	flags.StringVar(&publishDir, "publish-dir", "", "")
	flags.StringVar(&gitRepo, "git-repo", "", "")
	flags.StringVar(&gitToken, "git-token-path", "", "")
	flags.StringVar(&webhookSec, "webhook-secret", "", "")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		patch := map[string]any{}
		if desc != "" { patch["description"] = desc }
		if branch != "" { patch["branch"] = branch }
		if domain != "" { patch["domain"] = domain }
		if port != 0 { patch["binding_port"] = port }
		if service != "" { patch["service_name"] = service }
		if buildCmd != "" { patch["build_command"] = buildCmd }
		if runCmd != "" { patch["run_command"] = runCmd }
		if publishDir != "" { patch["publish_dir"] = publishDir }
		if gitRepo != "" { patch["git_repo_url"] = gitRepo }
		if gitToken != "" { patch["git_token_path"] = gitToken }
		if webhookSec != "" { patch["webhook_secret"] = webhookSec }

		if len(patch) == 0 {
			return fmt.Errorf("no fields specified; use flags to set fields (--help for list)")
		}
		var updated db.App
		if err := apiPatch("/apps/"+args[0], patch, &updated); err != nil {
			return err
		}
		fmt.Printf("✓ Updated %q\n", updated.Name)
		return nil
	}
	return cmd
}

func newAppsRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Deregister an app (does not delete files or stop services)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !promptYN(fmt.Sprintf("Remove %q from appctl?", args[0])) {
				fmt.Println("Aborted.")
				return nil
			}
			if err := apiDelete("/apps/" + args[0]); err != nil {
				return err
			}
			fmt.Printf("✓ Removed %q\n", args[0])
			return nil
		},
	}
}

func newAppsUnitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unit <name>",
		Short: "Write the systemd unit file for an app",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// The unit file writer needs the app data and calls deploy.WriteUnit.
			// Since we're in the CLI (not daemon), we open the local DB directly.
			dbPath := db.DefaultPath()
			database, err := db.Open(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer database.Close()

			app, err := db.GetApp(database, args[0])
			if err != nil || app == nil {
				return fmt.Errorf("app %q not found", args[0])
			}
			if err := deploy.WriteUnit(app); err != nil {
				return err
			}
			if err := deploy.EnableUnit(app.ServiceName); err != nil {
				return fmt.Errorf("enable unit: %w", err)
			}
			fmt.Printf("✓ Unit file written and enabled: /etc/systemd/system/%s.service\n", app.ServiceName)
			return nil
		},
	}
}
