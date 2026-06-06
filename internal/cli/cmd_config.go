package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View or update appctl daemon configuration",
	}
	cmd.AddCommand(newConfigShowCmd(), newConfigSetCmd())
	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Print all config values",
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg map[string]string
			if err := apiGet("/config", &cfg); err != nil {
				return err
			}
			tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(tw, "KEY\tVALUE")
			for k, v := range cfg {
				fmt.Fprintf(tw, "%s\t%s\n", k, v)
			}
			return tw.Flush()
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var cfg map[string]string
			if err := apiPatch("/config", map[string]string{args[0]: args[1]}, &cfg); err != nil {
				return err
			}
			fmt.Printf("✓ %s = %s\n", args[0], cfg[args[0]])
			return nil
		},
	}
}
