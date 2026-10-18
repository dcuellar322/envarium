// Package domain owns the canonical project manifest model and invariants.
package domain

import (
	"regexp"
)

const (
	// SchemaVersion is the stable v1 portable manifest contract.
	SchemaVersion = "envarium.dev/v1"
	// KindProject identifies the portable project manifest document.
	KindProject = "Project"
)

var manifestID = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// Manifest is the canonical portable project declaration.
type Manifest struct {
	SchemaVersion  string         `json:"schemaVersion" yaml:"schemaVersion" jsonschema:"required,enum=envarium.dev/v1"`
	Kind           string         `json:"kind" yaml:"kind" jsonschema:"required,enum=Project"`
	Metadata       Metadata       `json:"metadata" yaml:"metadata" jsonschema:"required"`
	Repository     Repository     `json:"repository" yaml:"repository" jsonschema:"required"`
	Runtime        Runtime        `json:"runtime,omitempty" yaml:"runtime,omitempty"`
	Lifecycle      Lifecycle      `json:"lifecycle,omitempty" yaml:"lifecycle,omitempty"`
	Services       []Service      `json:"services,omitempty" yaml:"services,omitempty"`
	Ports          []Port         `json:"ports,omitempty" yaml:"ports,omitempty"`
	Endpoints      []Endpoint     `json:"endpoints,omitempty" yaml:"endpoints,omitempty"`
	Actions        []Action       `json:"actions,omitempty" yaml:"actions,omitempty"`
	ResourcePolicy ResourcePolicy `json:"resourcePolicy,omitempty" yaml:"resourcePolicy,omitempty"`
}

// Metadata identifies a project for humans and automation.
type Metadata struct {
	ID          string   `json:"id" yaml:"id" jsonschema:"required,pattern=^[a-z0-9][a-z0-9-]*$,maxLength=63"`
	Name        string   `json:"name" yaml:"name" jsonschema:"required,minLength=1,maxLength=120"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty" jsonschema:"maxLength=1000"`
	Tags        []string `json:"tags,omitempty" yaml:"tags,omitempty"`
}

// Repository declares portable repository facts.
type Repository struct {
	Root          string `json:"root" yaml:"root" jsonschema:"required"`
	DefaultBranch string `json:"defaultBranch,omitempty" yaml:"defaultBranch,omitempty"`
}

// Runtime selects the project lifecycle driver.
type Runtime struct {
	Driver  string         `json:"driver,omitempty" yaml:"driver,omitempty" jsonschema:"enum=compose,enum=process,enum=external"`
	Compose *ComposeConfig `json:"compose,omitempty" yaml:"compose,omitempty"`
	Process *ProcessConfig `json:"process,omitempty" yaml:"process,omitempty"`
}

// ComposeConfig identifies portable Compose inputs.
type ComposeConfig struct {
	Files       []string `json:"files" yaml:"files" jsonschema:"required,minItems=1"`
	ProjectName string   `json:"projectName,omitempty" yaml:"projectName,omitempty"`
	Context     string   `json:"context,omitempty" yaml:"context,omitempty"`
	Profiles    []string `json:"profiles,omitempty" yaml:"profiles,omitempty" jsonschema:"maxItems=32,uniqueItems=true"`
}

// ProcessConfig declares project-wide environment and native process definitions.
type ProcessConfig struct {
	Environment map[string]string    `json:"environment,omitempty" yaml:"environment,omitempty"`
	Secrets     map[string]SecretRef `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	Processes   []ProcessDefinition  `json:"processes" yaml:"processes" jsonschema:"required,minItems=1"`
}

// ProcessDefinition is one shell-free native service command.
type ProcessDefinition struct {
	ID                 string               `json:"id" yaml:"id" jsonschema:"required"`
	Command            []string             `json:"command" yaml:"command" jsonschema:"required,minItems=1"`
	WorkingDirectory   string               `json:"workingDirectory,omitempty" yaml:"workingDirectory,omitempty"`
	Shell              bool                 `json:"shell,omitempty" yaml:"shell,omitempty"`
	Environment        map[string]string    `json:"environment,omitempty" yaml:"environment,omitempty"`
	Secrets            map[string]SecretRef `json:"secrets,omitempty" yaml:"secrets,omitempty"`
	Restart            RestartPolicy        `json:"restart,omitempty" yaml:"restart,omitempty"`
	StopTimeoutSeconds int                  `json:"stopTimeoutSeconds,omitempty" yaml:"stopTimeoutSeconds,omitempty" jsonschema:"minimum=0,maximum=300"`
}

// SecretRef identifies a value in an operating-system credential store without embedding it.
type SecretRef struct {
	Provider string `json:"provider" yaml:"provider" jsonschema:"required,enum=keychain"`
	Key      string `json:"key" yaml:"key" jsonschema:"required"`
	Account  string `json:"account,omitempty" yaml:"account,omitempty"`
}

// RestartPolicy enables bounded crash restart only when explicitly requested.
type RestartPolicy struct {
	Mode           string `json:"mode,omitempty" yaml:"mode,omitempty" jsonschema:"enum=never,enum=on-failure"`
	MaxRetries     int    `json:"maxRetries,omitempty" yaml:"maxRetries,omitempty" jsonschema:"minimum=0,maximum=20"`
	BackoffSeconds int    `json:"backoffSeconds,omitempty" yaml:"backoffSeconds,omitempty" jsonschema:"minimum=0,maximum=300"`
}

// Lifecycle declares the standard project mutations.
type Lifecycle struct {
	Start    *Invocation `json:"start,omitempty" yaml:"start,omitempty"`
	Stop     *Invocation `json:"stop,omitempty" yaml:"stop,omitempty"`
	Restart  *Invocation `json:"restart,omitempty" yaml:"restart,omitempty"`
	Rebuild  *Invocation `json:"rebuild,omitempty" yaml:"rebuild,omitempty"`
	Teardown *Invocation `json:"teardown,omitempty" yaml:"teardown,omitempty"`
}

// Invocation references an approved typed action.
type Invocation struct {
	Action  string         `json:"action" yaml:"action" jsonschema:"required"`
	Risk    string         `json:"risk,omitempty" yaml:"risk,omitempty" jsonschema:"enum=safe,enum=caution,enum=destructive"`
	Options map[string]any `json:"options,omitempty" yaml:"options,omitempty"`
}

// Service is a project-visible runtime unit.
type Service struct {
	ID           string        `json:"id" yaml:"id" jsonschema:"required"`
	DisplayName  string        `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Source       ServiceSource `json:"source" yaml:"source" jsonschema:"required"`
	Dependencies []string      `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	HealthChecks []HealthCheck `json:"healthChecks,omitempty" yaml:"healthChecks,omitempty"`
}

// ServiceSource binds a service to one runtime declaration.
type ServiceSource struct {
	ComposeService string `json:"composeService,omitempty" yaml:"composeService,omitempty"`
	Process        string `json:"process,omitempty" yaml:"process,omitempty"`
}

// HealthCheck declares bounded readiness validation.
type HealthCheck struct {
	ID                  string   `json:"id,omitempty" yaml:"id,omitempty"`
	Type                string   `json:"type" yaml:"type" jsonschema:"required,enum=http,enum=tcp,enum=process,enum=docker,enum=command,enum=composite"`
	URL                 string   `json:"url,omitempty" yaml:"url,omitempty"`
	Address             string   `json:"address,omitempty" yaml:"address,omitempty"`
	ExpectedStatus      int      `json:"expectedStatus,omitempty" yaml:"expectedStatus,omitempty" jsonschema:"minimum=0,maximum=599"`
	JSONPath            string   `json:"jsonPath,omitempty" yaml:"jsonPath,omitempty"`
	ExpectedValue       string   `json:"expectedValue,omitempty" yaml:"expectedValue,omitempty"`
	Command             []string `json:"command,omitempty" yaml:"command,omitempty"`
	Members             []string `json:"members,omitempty" yaml:"members,omitempty"`
	Mode                string   `json:"mode,omitempty" yaml:"mode,omitempty" jsonschema:"enum=all,enum=any"`
	InitialDelaySeconds int      `json:"initialDelaySeconds,omitempty" yaml:"initialDelaySeconds,omitempty" jsonschema:"minimum=0,maximum=3600"`
	IntervalSeconds     int      `json:"intervalSeconds,omitempty" yaml:"intervalSeconds,omitempty" jsonschema:"minimum=0,maximum=3600"`
	TimeoutSeconds      int      `json:"timeoutSeconds,omitempty" yaml:"timeoutSeconds,omitempty" jsonschema:"minimum=0,maximum=300"`
	Retries             int      `json:"retries,omitempty" yaml:"retries,omitempty" jsonschema:"minimum=0,maximum=20"`
	Severity            string   `json:"severity,omitempty" yaml:"severity,omitempty" jsonschema:"enum=info,enum=warning,enum=critical"`
	Required            bool     `json:"required,omitempty" yaml:"required,omitempty"`
}

// Port is a portable port declaration, not an observed binding.
type Port struct {
	ID       string `json:"id" yaml:"id" jsonschema:"required"`
	Service  string `json:"service,omitempty" yaml:"service,omitempty"`
	Host     int    `json:"host" yaml:"host" jsonschema:"required,minimum=1,maximum=65535"`
	Target   int    `json:"target" yaml:"target" jsonschema:"required,minimum=1,maximum=65535"`
	Protocol string `json:"protocol" yaml:"protocol" jsonschema:"required,enum=tcp,enum=udp"`
}

// Endpoint is a user-facing local URL.
type Endpoint struct {
	ID      string `json:"id" yaml:"id" jsonschema:"required"`
	Name    string `json:"name" yaml:"name" jsonschema:"required"`
	URL     string `json:"url" yaml:"url" jsonschema:"required"`
	Primary bool   `json:"primary,omitempty" yaml:"primary,omitempty"`
}

// Action is an explicit argument-array command or typed built-in capability.
type Action struct {
	ID               string            `json:"id" yaml:"id" jsonschema:"required"`
	Name             string            `json:"name" yaml:"name" jsonschema:"required"`
	Type             string            `json:"type" yaml:"type" jsonschema:"required,enum=terminal.open,enum=editor.open,enum=browser.open,enum=command,enum=command.run,enum=agent.start,enum=git.fetch,enum=git.pull,enum=git.push,enum=tests.run,enum=migration.run"`
	Command          []string          `json:"command,omitempty" yaml:"command,omitempty"`
	WorkingDirectory string            `json:"workingDirectory,omitempty" yaml:"workingDirectory,omitempty"`
	Shell            bool              `json:"shell,omitempty" yaml:"shell,omitempty"`
	CaptureOutput    bool              `json:"captureOutput,omitempty" yaml:"captureOutput,omitempty"`
	Provider         string            `json:"provider,omitempty" yaml:"provider,omitempty"`
	Target           string            `json:"target,omitempty" yaml:"target,omitempty"`
	Risk             string            `json:"risk,omitempty" yaml:"risk,omitempty" jsonschema:"enum=read_only,enum=mutating,enum=networked,enum=destructive,enum=interactive"`
	TimeoutSeconds   int               `json:"timeoutSeconds,omitempty" yaml:"timeoutSeconds,omitempty" jsonschema:"minimum=0,maximum=86400"`
	Environment      map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
}

// ResourcePolicy contains warning thresholds only.
type ResourcePolicy struct {
	Warnings ResourceWarnings `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

// ResourceWarnings are advisory local thresholds.
type ResourceWarnings struct {
	MemoryMiB  int `json:"memoryMiB,omitempty" yaml:"memoryMiB,omitempty" jsonschema:"minimum=0"`
	CPUPercent int `json:"cpuPercent,omitempty" yaml:"cpuPercent,omitempty" jsonschema:"minimum=0,maximum=100"`
	StorageGiB int `json:"storageGiB,omitempty" yaml:"storageGiB,omitempty" jsonschema:"minimum=0"`
}
