package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <name>",
		Short: "Start an app's systemd service",
		Args:  cobra.ExactArgs(1),
		RunE:  serviceAction("start"),
	}
}

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop <name>",
		Short: "Stop an app's systemd service",
		Args:  cobra.ExactArgs(1),
		RunE:  serviceAction("stop"),
	}
}

func newRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart <name>",
		Short: "Restart an app's systemd service",
		Args:  cobra.ExactArgs(1),
		RunE:  serviceAction("restart"),
	}
}

func serviceAction(action string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := apiPost("/apps/"+args[0]+"/"+action, nil, nil); err != nil {
			return err
		}
		fmt.Printf("✓ %s %q\n", action, args[0])
		return nil
	}
}
