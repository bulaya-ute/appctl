package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/bulaya-ute/appctl/internal/db"
	"github.com/bulaya-ute/appctl/internal/deploy"
)

func (s *Server) handleListApps(w http.ResponseWriter, r *http.Request) {
	apps, err := db.ListApps(s.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if apps == nil {
		apps = []db.App{}
	}
	writeJSON(w, http.StatusOK, apps)
}

func (s *Server) handleGetApp(w http.ResponseWriter, r *http.Request) {
	app, err := db.GetApp(s.db, chi.URLParam(r, "name"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if app == nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	writeJSON(w, http.StatusOK, app)
}

func (s *Server) handleCreateApp(w http.ResponseWriter, r *http.Request) {
	var a db.App
	if err := decodeJSON(r, &a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if a.Name == "" || a.Type == "" || a.LocalPath == "" {
		writeError(w, http.StatusBadRequest, "name, type, and local_path are required")
		return
	}
	if a.Branch == "" {
		a.Branch = "main"
	}
	if a.Source == "" {
		a.Source = db.SourceGit
	}
	if err := db.CreateApp(s.db, &a); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (s *Server) handleUpdateApp(w http.ResponseWriter, r *http.Request) {
	existing, err := db.GetApp(s.db, chi.URLParam(r, "name"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	// Decode partial update into existing struct — unset JSON fields keep their zero values,
	// so callers should send only the fields they want to change.
	var patch db.App
	if err := decodeJSON(r, &patch); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	applyPatch(existing, &patch)
	if err := db.UpdateApp(s.db, existing); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, existing)
}

func (s *Server) handleDeleteApp(w http.ResponseWriter, r *http.Request) {
	app, err := db.GetApp(s.db, chi.URLParam(r, "name"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if app == nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if err := db.DeleteApp(s.db, app.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleDeploy(w http.ResponseWriter, r *http.Request) {
	app, err := db.GetApp(s.db, chi.URLParam(r, "name"))
	if err != nil || app == nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}

	var body struct {
		Version string `json:"version"`
	}
	_ = decodeJSON(r, &body)
	if body.Version == "" {
		body.Version = "latest"
	}

	dep := &db.Deployment{
		AppID:       app.ID,
		Version:     body.Version,
		TriggeredBy: db.TriggerManual,
		Status:      db.DeployStatusRunning,
	}
	if err := db.CreateDeployment(s.db, dep); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	caddyURL := db.GetConfigValue(s.db, "caddy_admin_url", "http://localhost:2019")
	appCopy := *app

	go func() {
		result := deploy.Run(&appCopy, body.Version, caddyURL)
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

	writeJSON(w, http.StatusAccepted, dep)
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	s.serviceAction(w, r, deploy.StartService)
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	s.serviceAction(w, r, deploy.StopService)
}

func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	s.serviceAction(w, r, func(name string) error {
		return deploy.RestartService(name, func(string, ...any) {})
	})
}

func (s *Server) serviceAction(w http.ResponseWriter, r *http.Request, fn func(string) error) {
	app, err := db.GetApp(s.db, chi.URLParam(r, "name"))
	if err != nil || app == nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	if app.ServiceName == "" {
		writeError(w, http.StatusBadRequest, "app has no service_name configured")
		return
	}
	if err := fn(app.ServiceName); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleListDeployments(w http.ResponseWriter, r *http.Request) {
	app, err := db.GetApp(s.db, chi.URLParam(r, "name"))
	if err != nil || app == nil {
		writeError(w, http.StatusNotFound, "app not found")
		return
	}
	deps, err := db.ListDeployments(s.db, app.ID, 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if deps == nil {
		deps = []db.Deployment{}
	}
	writeJSON(w, http.StatusOK, deps)
}

// applyPatch copies non-zero fields from patch onto target.
func applyPatch(target, patch *db.App) {
	if patch.Description != "" {
		target.Description = patch.Description
	}
	if patch.Type != "" {
		target.Type = patch.Type
	}
	if patch.Source != "" {
		target.Source = patch.Source
	}
	if patch.LocalPath != "" {
		target.LocalPath = patch.LocalPath
	}
	if patch.GitRepoURL != "" {
		target.GitRepoURL = patch.GitRepoURL
	}
	if patch.GitTokenPath != "" {
		target.GitTokenPath = patch.GitTokenPath
	}
	if patch.Branch != "" {
		target.Branch = patch.Branch
	}
	if patch.ServiceName != "" {
		target.ServiceName = patch.ServiceName
	}
	if patch.BindingPort != 0 {
		target.BindingPort = patch.BindingPort
	}
	if patch.Domain != "" {
		target.Domain = patch.Domain
	}
	if patch.BuildCommand != "" {
		target.BuildCommand = patch.BuildCommand
	}
	if patch.RunCommand != "" {
		target.RunCommand = patch.RunCommand
	}
	if patch.PublishDir != "" {
		target.PublishDir = patch.PublishDir
	}
	if patch.WebhookSecret != "" {
		target.WebhookSecret = patch.WebhookSecret
	}
}
