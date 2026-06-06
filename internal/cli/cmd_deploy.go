package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/bulaya-ute/appctl/internal/db"
)

func newDeployCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deploy <name> [version]",
		Short: "Deploy an app (optionally to a specific version/tag)",
		Long: `Deploy an app by pulling from git, building, restarting its service,
and updating its Caddy route.

Version can be a git tag (e.g. "0.2.0" or "v0.2.0") or omitted / "latest" to
check out the configured branch HEAD.`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := "latest"
			if len(args) == 2 {
				version = args[1]
			}
			var dep db.Deployment
			if err := apiPost("/apps/"+args[0]+"/deploy",
				map[string]string{"version": version}, &dep); err != nil {
				return err
			}
			fmt.Printf("✓ Deploy accepted (id: %s, version: %s)\n", dep.ID, dep.Version)
			fmt.Printf("  Run 'appctl logs %s' to see the deployment log.\n", args[0])
			return nil
		},
	}
}
