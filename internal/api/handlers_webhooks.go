package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/bulaya-ute/appctl/internal/db"
	"github.com/bulaya-ute/appctl/internal/deploy"
)

func (s *Server) handleGitHubWebhook(w http.ResponseWriter, r *http.Request) {
	appName := chi.URLParam(r, "name")
	app, err := db.GetApp(s.db, appName)
	if err != nil || app == nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cannot read body")
		return
	}

	if !verifySignature(app.WebhookSecret, body, r.Header.Get("X-Hub-Signature-256")) {
		writeError(w, http.StatusUnauthorized, "invalid signature")
		return
	}

	// Only act on release.published events.
	if r.Header.Get("X-GitHub-Event") != "release" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	var payload struct {
		Action  string `json:"action"`
		Release struct {
			TagName string `json:"tag_name"`
		} `json:"release"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if payload.Action != "published" {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	version := strings.TrimPrefix(payload.Release.TagName, "v")

	dep := &db.Deployment{
		AppID:       app.ID,
		Version:     version,
		TriggeredBy: db.TriggerWebhook,
		Status:      db.DeployStatusRunning,
	}
	if err := db.CreateDeployment(s.db, dep); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	caddyURL := db.GetConfigValue(s.db, "caddy_admin_url", "http://localhost:2019")
	appCopy := *app

	go func() {
		result := deploy.Run(&appCopy, version, caddyURL)
		now := time.Now().UTC()
		dep.FinishedAt = &now
		dep.Log = result.Log
		if result.Err != nil {
			dep.Status = db.DeployStatusFailed
			dep.Log += "\nERROR: " + result.Err.Error()
		} else {
			dep.Status = db.DeployStatusSuccess
		}
		_ = db.UpdateDeployment(s.db, dep)
	}()

	writeJSON(w, http.StatusAccepted, map[string]string{
		"status":  "accepted",
		"version": version,
	})
}

func verifySignature(secret string, body []byte, header string) bool {
	if secret == "" {
		return true
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(header))
}
