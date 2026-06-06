package db

import "time"

type AppType string

const (
	AppTypeDotnetAPI AppType = "dotnet-api"
	AppTypeReactSPA  AppType = "react-spa"
	AppTypeStatic    AppType = "static"
	AppTypeCustom    AppType = "custom"
)

type SourceType string

const (
	SourceGit   SourceType = "git"
	SourceLocal SourceType = "local"
)

type DeployStatus string

const (
	DeployStatusPending DeployStatus = "pending"
	DeployStatusRunning DeployStatus = "running"
	DeployStatusSuccess DeployStatus = "success"
	DeployStatusFailed  DeployStatus = "failed"
)

type TriggerType string

const (
	TriggerManual  TriggerType = "manual"
	TriggerWebhook TriggerType = "webhook"
)

// App is a managed application registered with appctl.
type App struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Description  string     `json:"description"`
	Type         AppType    `json:"type"`
	Source       SourceType `json:"source"`
	LocalPath    string     `json:"local_path"`
	GitRepoURL   string     `json:"git_repo_url"`
	GitTokenPath string     `json:"git_token_path"` // path to file containing token
	Branch       string     `json:"branch"`
	ServiceName  string     `json:"service_name"`  // systemd service name
	BindingPort  int        `json:"binding_port"`  // upstream port for Caddy reverse proxy
	Domain       string     `json:"domain"`        // domain to configure in Caddy
	BuildCommand string     `json:"build_command"` // overrides default for the type
	RunCommand   string     `json:"run_command"`   // overrides ExecStart in systemd unit
	PublishDir   string     `json:"publish_dir"`   // output dir; default derived per type
	WebhookSecret string    `json:"webhook_secret,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Deployment is a single deploy event for an app.
type Deployment struct {
	ID          string       `json:"id"`
	AppID       string       `json:"app_id"`
	AppName     string       `json:"app_name,omitempty"` // populated via join
	Version     string       `json:"version"`
	TriggeredBy TriggerType  `json:"triggered_by"`
	Status      DeployStatus `json:"status"`
	Log         string       `json:"log"`
	StartedAt   time.Time    `json:"started_at"`
	FinishedAt  *time.Time   `json:"finished_at"`
}
