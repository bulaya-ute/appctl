package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bulaya-ute/appctl/internal/db"
	"github.com/bulaya-ute/appctl/internal/export"
)

func newExportCmd() *cobra.Command {
	var outFile string
	cmd := &cobra.Command{
		Use:   "export setup [app1 app2 ...]",
		Short: "Generate a setup.sh pre-configured for a set of apps",
		Long: `Generates a self-contained setup.sh for the given apps (or all registered apps
if none are specified). The script can be committed to a GitHub repo as
deploy/setup.sh and run with a one-liner curl command.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var apps []db.App
			if err := apiGet("/apps", &apps); err != nil {
				return err
			}

			// Filter to requested apps if names provided.
			if len(args) > 0 {
				want := make(map[string]bool, len(args))
				for _, n := range args {
					want[n] = true
				}
				var filtered []db.App
				for _, a := range apps {
					if want[a.Name] {
						filtered = append(filtered, a)
					}
				}
				if len(filtered) == 0 {
					return fmt.Errorf("none of the specified apps are registered")
				}
				apps = filtered
			}

			if len(apps) == 0 {
				return fmt.Errorf("no apps registered; add some with 'appctl apps add'")
			}

			script, err := export.GenerateSetup(apps)
			if err != nil {
				return err
			}

			if outFile == "" || outFile == "-" {
				fmt.Print(script)
				return nil
			}

			if err := os.WriteFile(outFile, []byte(script), 0o755); err != nil {
				return err
			}
			fmt.Printf("✓ Written to %s\n", outFile)
			return nil
		},
	}
	cmd.Flags().StringVarP(&outFile, "out", "o", "-", "output file (default: stdout)")
	return cmd
}
