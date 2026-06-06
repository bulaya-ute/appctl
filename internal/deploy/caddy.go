package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/bulaya-ute/appctl/internal/db"
)

// UpsertCaddyRoute adds or updates the Caddy reverse proxy / file_server route for an app.
// Routes are tagged with @id "appctl-<name>" so subsequent deploys can update in place.
func UpsertCaddyRoute(app *db.App, adminURL string, logf func(string, ...any)) error {
	adminURL = strings.TrimRight(adminURL, "/")
	routeID := "appctl-" + app.Name

	var route map[string]any
	if app.Type == db.AppTypeReactSPA || app.Type == db.AppTypeStatic {
		root := app.PublishDir
		if root == "" {
			root = filepath.Join(app.LocalPath, "dist")
		}
		route = fileServerRoute(app.Domain, root)
	} else {
		if app.BindingPort == 0 {
			logf("warn: no binding_port set, skipping caddy route")
			return nil
		}
		route = reverseProxyRoute(app.Domain, app.BindingPort)
	}
	route["@id"] = routeID

	body, err := json.Marshal(route)
	if err != nil {
		return err
	}

	// Try PATCH /id/<routeID> first (update existing).
	if err := caddyDo(http.MethodPatch, adminURL+"/id/"+routeID, body); err == nil {
		logf("caddy route updated: %s → %s", app.Domain, routeID)
		return nil
	}

	// Route not found; append to the server's route list.
	appendURL := adminURL + "/config/apps/http/servers/srv0/routes/..."
	if err := caddyDo(http.MethodPost, appendURL, body); err != nil {
		return fmt.Errorf("caddy append route: %w", err)
	}
	logf("caddy route created: %s → %s", app.Domain, routeID)
	return nil
}

// RemoveCaddyRoute removes the Caddy route for an app.
func RemoveCaddyRoute(appName, adminURL string) error {
	adminURL = strings.TrimRight(adminURL, "/")
	routeID := "appctl-" + appName
	return caddyDo(http.MethodDelete, adminURL+"/id/"+routeID, nil)
}

func caddyDo(method, url string, body []byte) error {
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("caddy %s %s → %d: %s", method, url, resp.StatusCode, strings.TrimSpace(string(msg)))
	}
	return nil
}

func reverseProxyRoute(domain string, port int) map[string]any {
	return map[string]any{
		"match": []map[string]any{{"host": []string{domain}}},
		"handle": []map[string]any{{
			"handler":   "reverse_proxy",
			"upstreams": []map[string]any{{"dial": fmt.Sprintf("localhost:%d", port)}},
		}},
	}
}

func fileServerRoute(domain, root string) map[string]any {
	return map[string]any{
		"match": []map[string]any{{"host": []string{domain}}},
		"handle": []map[string]any{{
			"handler": "file_server",
			"root":    root,
		}},
	}
}
