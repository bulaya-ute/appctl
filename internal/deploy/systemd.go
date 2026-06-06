package deploy

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/bulaya-ute/appctl/internal/db"
)

const unitTmpl = `[Unit]
Description=appctl: {{.Description}}
After=network.target

[Service]
Type=simple
WorkingDirectory={{.WorkingDir}}
ExecStart={{.ExecStart}}
Restart=always
RestartSec=10
{{- range .Env}}
Environment={{.}}
{{- end}}

[Install]
WantedBy=multi-user.target
`

type unitData struct {
	Description string
	WorkingDir  string
	ExecStart   string
	Env         []string
}

// WriteUnit creates or overwrites /etc/systemd/system/<service>.service.
func WriteUnit(app *db.App) error {
	if app.ServiceName == "" {
		return fmt.Errorf("app %q has no service_name", app.Name)
	}
	execStart, err := resolveExecStart(app)
	if err != nil {
		return err
	}

	workDir := app.LocalPath
	if app.Type == db.AppTypeDotnetAPI {
		workDir = resolvePublishDir(app)
	}

	data := unitData{
		Description: app.Name,
		WorkingDir:  workDir,
		ExecStart:   execStart,
	}
	if app.BindingPort > 0 && app.Type == db.AppTypeDotnetAPI {
		data.Env = append(data.Env,
			fmt.Sprintf("ASPNETCORE_URLS=http://localhost:%d", app.BindingPort),
			"ASPNETCORE_ENVIRONMENT=Production",
		)
	}

	var buf bytes.Buffer
	tmpl := template.Must(template.New("unit").Parse(unitTmpl))
	if err := tmpl.Execute(&buf, data); err != nil {
		return err
	}

	unitPath := fmt.Sprintf("/etc/systemd/system/%s.service", app.ServiceName)
	return os.WriteFile(unitPath, buf.Bytes(), 0o644)
}

// EnableUnit runs systemctl daemon-reload + enable for the service.
func EnableUnit(serviceName string) error {
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}
	return exec.Command("systemctl", "enable", serviceName).Run()
}

// RestartService reloads systemd and restarts the named service.
func RestartService(name string, logf func(string, ...any)) error {
	logf("reloading systemd daemon")
	if err := run(nil, logf, "systemctl", "daemon-reload"); err != nil {
		return err
	}
	logf("restarting service %s", name)
	return run(nil, logf, "systemctl", "restart", name)
}

// StartService starts the named systemd service.
func StartService(name string) error {
	return exec.Command("systemctl", "start", name).Run()
}

// StopService stops the named systemd service.
func StopService(name string) error {
	return exec.Command("systemctl", "stop", name).Run()
}

func resolveExecStart(app *db.App) (string, error) {
	if app.RunCommand != "" {
		return app.RunCommand, nil
	}
	switch app.Type {
	case db.AppTypeDotnetAPI:
		publishDir := resolvePublishDir(app)
		dll, err := findDotnetDLL(publishDir)
		if err != nil {
			return "", fmt.Errorf(
				"cannot determine ExecStart for %q: %w\nHint: set run_command or deploy the app first",
				app.Name, err,
			)
		}
		return "/usr/bin/dotnet " + dll, nil
	default:
		return "", fmt.Errorf(
			"cannot derive ExecStart for type %q; set run_command explicitly",
			app.Type,
		)
	}
}

func resolvePublishDir(app *db.App) string {
	if app.PublishDir != "" {
		return app.PublishDir
	}
	return filepath.Join(app.LocalPath, "publish")
}

func findDotnetDLL(publishDir string) (string, error) {
	entries, err := os.ReadDir(publishDir)
	if err != nil {
		return "", fmt.Errorf("publish dir not found: %s", publishDir)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".runtimeconfig.json") {
			name := strings.TrimSuffix(e.Name(), ".runtimeconfig.json")
			return filepath.Join(publishDir, name+".dll"), nil
		}
	}
	return "", fmt.Errorf("no .dll found in %s", publishDir)
}
