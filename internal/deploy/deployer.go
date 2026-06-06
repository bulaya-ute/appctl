package deploy

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bulaya-ute/appctl/internal/db"
)

// Result holds the captured log and final error from a deploy run.
type Result struct {
	Log string
	Err error
}

// Run executes the full deploy pipeline for an app:
// git pull → build → service restart → caddy route update.
func Run(app *db.App, version, caddyAdminURL string) Result {
	var buf bytes.Buffer
	logf := func(format string, args ...any) {
		fmt.Fprintf(&buf, format+"\n", args...)
	}

	if err := pull(app, version, logf); err != nil {
		return Result{Log: buf.String(), Err: err}
	}
	if err := build(app, logf); err != nil {
		return Result{Log: buf.String(), Err: err}
	}
	if app.ServiceName != "" {
		if err := RestartService(app.ServiceName, logf); err != nil {
			return Result{Log: buf.String(), Err: fmt.Errorf("service restart: %w", err)}
		}
	}
	if app.Domain != "" && caddyAdminURL != "" {
		if err := UpsertCaddyRoute(app, caddyAdminURL, logf); err != nil {
			logf("warn: caddy update failed: %v", err)
		}
	}
	return Result{Log: buf.String()}
}

func pull(app *db.App, version string, logf func(string, ...any)) error {
	if app.Source == db.SourceLocal {
		logf("source=local, skipping git operations")
		return nil
	}
	gitDir := filepath.Join(app.LocalPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		logf("cloning %s → %s", app.GitRepoURL, app.LocalPath)
		cloneArgs := []string{"clone"}
		if token := readToken(app.GitTokenPath); token != "" {
			// inject token into URL for HTTPS
			repoURL := injectToken(app.GitRepoURL, token)
			cloneArgs = append(cloneArgs, repoURL, app.LocalPath)
		} else {
			cloneArgs = append(cloneArgs, app.GitRepoURL, app.LocalPath)
		}
		if err := run(&bytes.Buffer{}, logf, "git", cloneArgs...); err != nil {
			return err
		}
	}

	if err := run(nil, logf, "git", "-C", app.LocalPath, "fetch", "--tags", "--prune"); err != nil {
		return err
	}

	ref := resolveRef(version, app.Branch)
	logf("checking out %s", ref)
	return run(nil, logf, "git", "-C", app.LocalPath, "-c", "advice.detachedHead=false", "checkout", ref)
}

func resolveRef(version, branch string) string {
	if version == "" || version == "latest" {
		return branch
	}
	v := strings.TrimPrefix(version, "v")
	return "tags/v" + v
}

func build(app *db.App, logf func(string, ...any)) error {
	switch app.Type {
	case db.AppTypeDotnetAPI:
		return buildDotnet(app, logf)
	case db.AppTypeReactSPA, db.AppTypeStatic:
		return buildNode(app, logf)
	case db.AppTypeCustom:
		return buildCustom(app, logf)
	}
	return nil
}

func buildDotnet(app *db.App, logf func(string, ...any)) error {
	if app.BuildCommand != "" {
		return runShell(logf, app.BuildCommand, app.LocalPath)
	}
	publishDir := app.PublishDir
	if publishDir == "" {
		publishDir = filepath.Join(app.LocalPath, "publish")
	}
	return run(nil, logf, "dotnet", "publish", app.LocalPath, "-c", "Release", "-o", publishDir)
}

func buildNode(app *db.App, logf func(string, ...any)) error {
	if app.BuildCommand != "" {
		return runShell(logf, app.BuildCommand, app.LocalPath)
	}
	if err := run(nil, logf, "npm", "ci", "--prefix", app.LocalPath); err != nil {
		return err
	}
	return run(nil, logf, "npm", "--prefix", app.LocalPath, "run", "build")
}

func buildCustom(app *db.App, logf func(string, ...any)) error {
	if app.BuildCommand == "" {
		logf("type=custom but no build_command set, skipping build")
		return nil
	}
	return runShell(logf, app.BuildCommand, app.LocalPath)
}

func run(extraBuf *bytes.Buffer, logf func(string, ...any), name string, args ...string) error {
	logf("$ %s %s", name, strings.Join(args, " "))
	cmd := exec.Command(name, args...)
	cmd.Stdout = &multiWriter{logf: logf, extra: extraBuf}
	cmd.Stderr = &multiWriter{logf: logf, extra: extraBuf}
	return cmd.Run()
}

func runShell(logf func(string, ...any), command, dir string) error {
	logf("$ sh -c %q", command)
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = &multiWriter{logf: logf}
	cmd.Stderr = &multiWriter{logf: logf}
	return cmd.Run()
}

type multiWriter struct {
	logf  func(string, ...any)
	extra *bytes.Buffer
}

func (w *multiWriter) Write(p []byte) (int, error) {
	line := strings.TrimRight(string(p), "\n")
	if line != "" {
		w.logf("%s", line)
	}
	if w.extra != nil {
		w.extra.Write(p)
	}
	return len(p), nil
}

func readToken(path string) string {
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func injectToken(repoURL, token string) string {
	// https://github.com/... → https://<token>@github.com/...
	for _, prefix := range []string{"https://", "http://"} {
		if strings.HasPrefix(repoURL, prefix) {
			return prefix + token + "@" + repoURL[len(prefix):]
		}
	}
	return repoURL
}
