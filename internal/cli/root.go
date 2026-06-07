package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var daemonHost string

var rootCmd = &cobra.Command{
	Use:   "appctl",
	Short: "appctl — deployment manager for self-hosted applications",
	Long: `appctl manages deployments on a single server: git pull, build,
systemd service lifecycle, and Caddy reverse proxy configuration.

Run "appctl server" to start the daemon, then use subcommands to manage apps.`,
}

func Execute(version string) {
	rootCmd.Version = version
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&daemonHost, "host",
		envOr("APPCTL_HOST", "http://localhost:7070"),
		"appctl daemon address (env: APPCTL_HOST)")

	rootCmd.AddCommand(
		newServerCmd(),
		newAppsCmd(),
		newDeployCmd(),
		newStartCmd(),
		newStopCmd(),
		newRestartCmd(),
		newLogsCmd(),
		newConfigCmd(),
		newExportCmd(),
	)
}

// --- HTTP client helpers ---

func apiGet(path string, out any) error {
	resp, err := http.Get(daemonHost + path)
	if err != nil {
		return fmt.Errorf("cannot reach daemon at %s: %w\nIs appctl server running?", daemonHost, err)
	}
	defer resp.Body.Close()
	return checkAndDecode(resp, out)
}

func apiPost(path string, body, out any) error {
	return apiDo(http.MethodPost, path, body, out)
}

func apiPatch(path string, body, out any) error {
	return apiDo(http.MethodPatch, path, body, out)
}

func apiDelete(path string) error {
	return apiDo(http.MethodDelete, path, nil, nil)
}

func apiDo(method, path string, body, out any) error {
	var r io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		r = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, daemonHost+path, r)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach daemon at %s: %w\nIs appctl server running?", daemonHost, err)
	}
	defer resp.Body.Close()
	return checkAndDecode(resp, out)
}

func checkAndDecode(resp *http.Response, out any) error {
	if resp.StatusCode >= 400 {
		var e map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&e); err == nil {
			if msg, ok := e["error"]; ok {
				return fmt.Errorf("server error: %s", msg)
			}
		}
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}
	if out != nil && resp.StatusCode != http.StatusNoContent {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// --- Prompt helpers ---

func prompt(label, def string) string {
	if def != "" {
		fmt.Printf("  %s [%s]: ", label, def)
	} else {
		fmt.Printf("  %s: ", label)
	}
	line, _ := readLine()
	if line == "" {
		return def
	}
	return line
}

func promptRequired(label string) string {
	for {
		fmt.Printf("  %s (required): ", label)
		line, _ := readLine()
		if line != "" {
			return line
		}
		fmt.Println("  This field is required.")
	}
}

func promptChoice(label string, choices []string) string {
	fmt.Printf("  %s\n", label)
	for i, c := range choices {
		fmt.Printf("    %d) %s\n", i+1, c)
	}
	for {
		fmt.Printf("  Choice [1]: ")
		line, _ := readLine()
		if line == "" {
			return choices[0]
		}
		for i, c := range choices {
			if line == fmt.Sprintf("%d", i+1) {
				return c
			}
		}
		fmt.Println("  Invalid choice.")
	}
}

func promptYN(label string) bool {
	fmt.Printf("  %s [y/N]: ", label)
	line, _ := readLine()
	return strings.ToLower(strings.TrimSpace(line)) == "y"
}

func readLine() (string, error) {
	var buf []byte
	b := make([]byte, 1)
	for {
		_, err := os.Stdin.Read(b)
		if err != nil || b[0] == '\n' {
			break
		}
		buf = append(buf, b[0])
	}
	return strings.TrimRight(string(buf), "\r"), nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
