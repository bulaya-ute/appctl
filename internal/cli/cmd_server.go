package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/bulaya-ute/appctl/internal/api"
	"github.com/bulaya-ute/appctl/internal/db"
)

func newServerCmd() *cobra.Command {
	var dbPath string
	var port string

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start the appctl daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			if dbPath == "" {
				dbPath = db.DefaultPath()
			}

			database, err := db.Open(dbPath)
			if err != nil {
				return fmt.Errorf("open database: %w", err)
			}
			defer database.Close()

			if port == "" {
				port = db.GetConfigValue(database, "appctl_port", "7070")
			}
			addr := ":" + port

			fmt.Fprintf(os.Stderr, "appctl daemon listening on %s (db: %s)\n", addr, dbPath)
			return api.New(database).Run(addr)
		},
	}

	cmd.Flags().StringVar(&dbPath, "db", "", "path to SQLite database (default: "+db.DefaultPath()+")")
	cmd.Flags().StringVar(&port, "port", "", "port to listen on (default from db config: 7070)")
	return cmd
}
