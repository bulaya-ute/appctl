package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/bulaya-ute/appctl/internal/db"
)

func newLogsCmd() *cobra.Command {
	var showLog bool
	cmd := &cobra.Command{
		Use:   "logs <name>",
		Short: "Show deployment history for an app",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var deps []db.Deployment
			if err := apiGet("/apps/"+args[0]+"/deployments", &deps); err != nil {
				return err
			}
			if len(deps) == 0 {
				fmt.Printf("No deployments found for %q.\n", args[0])
				return nil
			}

			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "ID\tVERSION\tSTATUS\tTRIGGER\tSTARTED\tDURATION")
			for _, d := range deps {
				dur := "running"
				if d.FinishedAt != nil {
					dur = d.FinishedAt.Sub(d.StartedAt).Round(time.Second).String()
				}
				fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
					d.ID[:8], d.Version, d.Status, d.TriggeredBy,
					d.StartedAt.Local().Format("2006-01-02 15:04:05"), dur,
				)
			}
			tw.Flush()

			if showLog && len(deps) > 0 {
				fmt.Printf("\n── Log for last deployment ──\n%s\n", deps[0].Log)
			}
			return nil
		},
	}
	cmd.Flags().BoolVarP(&showLog, "log", "l", false, "print the full log of the most recent deployment")
	return cmd
}
