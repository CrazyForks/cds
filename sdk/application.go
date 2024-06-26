package sdk

import (
	"strings"
	"time"
)

// Repository structs contains all needed information about a single repository
type Repository struct {
	URL  string
	Hook bool
}

// Application represent an application in a project
type Application struct {
	ID                   int64                        `json:"id" db:"id"`
	Name                 string                       `json:"name" db:"name" cli:"name,key" action_metadata:"application-name"`
	Description          string                       `json:"description" db:"description"`
	Icon                 string                       `json:"icon" db:"icon"`
	ProjectID            int64                        `json:"-" db:"project_id"`
	ProjectKey           string                       `json:"project_key" db:"-" cli:"project_key"`
	Variables            []ApplicationVariable        `json:"variables,omitempty" db:"-"`
	Notifications        []UserNotification           `json:"notifications,omitempty" db:"-"`
	LastModified         time.Time                    `json:"last_modified" db:"last_modified" mapstructure:"-"`
	VCSServer            string                       `json:"vcs_server,omitempty" db:"vcs_server"`
	RepositoryFullname   string                       `json:"repository_fullname,omitempty" db:"repo_fullname" cli:"repository_fullname"`
	RepositoryStrategy   RepositoryStrategy           `json:"vcs_strategy,omitempty" db:"cipher_vcs_strategy" gorpmapping:"encrypted,ProjectID,Name"`
	Metadata             Metadata                     `json:"metadata" yaml:"metadata" db:"metadata"`
	Keys                 []ApplicationKey             `json:"keys" yaml:"keys" db:"-"`
	Usage                *Usage                       `json:"usage,omitempty" db:"-" cli:"-"`
	DeploymentStrategies map[string]IntegrationConfig `json:"deployment_strategies,omitempty" db:"-" cli:"-"`
	FromRepository       string                       `json:"from_repository,omitempty" db:"from_repository" cli:"-"`
	// aggregate
	WorkflowAscodeHolder *Workflow `json:"workflow_ascode_holder,omitempty" cli:"-" db:"-"`
}

// IsValid returns error if the application is not valid.
func (app Application) IsValid() error {
	if !NamePatternRegex.MatchString(app.Name) {
		return NewErrorFrom(ErrInvalidName, "application name should match pattern %s", NamePattern)
	}

	if app.Icon != "" {
		if !strings.HasPrefix(app.Icon, IconFormat) {
			return ErrIconBadFormat
		}
		if len(app.Icon) > MaxIconSize {
			return ErrIconBadSize
		}
	}

	return nil
}

// SSHKeys returns the slice of ssh key for an application
func (app Application) SSHKeys() []ApplicationKey {
	keys := []ApplicationKey{}
	for _, k := range app.Keys {
		if k.Type == KeyTypeSSH {
			keys = append(keys, k)
		}
	}
	return keys
}

// PGPKeys returns the slice of pgp key for an application
func (app Application) PGPKeys() []ApplicationKey {
	keys := []ApplicationKey{}
	for _, k := range app.Keys {
		if k.Type == KeyTypePGP {
			keys = append(keys, k)
		}
	}
	return keys
}

// RepositoryStrategy represents the way to use the repository
type RepositoryStrategy struct {
	ConnectionType string `json:"connection_type"`
	SSHKey         string `json:"ssh_key"`
	SSHKeyContent  string `json:"ssh_key_content,omitempty"`
	User           string `json:"user"`
	Password       string `json:"password"`
	Branch         string `json:"branch,omitempty"`
	DefaultBranch  string `json:"default_branch,omitempty"`
	PGPKey         string `json:"pgp_key"`
}

// ApplicationVariableAudit represents an audit on an application variable
type ApplicationVariableAudit struct {
	ID             int64                `json:"id" yaml:"-" db:"id"`
	ApplicationID  int64                `json:"application_id" yaml:"-" db:"application_id"`
	VariableID     int64                `json:"variable_id" yaml:"-" db:"variable_id"`
	Type           string               `json:"type" yaml:"-" db:"type"`
	VariableBefore *ApplicationVariable `json:"variable_before,omitempty" yaml:"-" db:"-"`
	VariableAfter  ApplicationVariable  `json:"variable_after,omitempty" yaml:"-" db:"-"`
	Versionned     time.Time            `json:"versionned" yaml:"-" db:"versionned"`
	Author         string               `json:"author" yaml:"-" db:"author"`
}

// GetKey return a key by name
func (app Application) GetKey(kname string) *ApplicationKey {
	for i := range app.Keys {
		if app.Keys[i].Name == kname {
			return &app.Keys[i]
		}
	}
	return nil
}

// GetSSHKey return a key by name
func (app Application) GetSSHKey(kname string) *ApplicationKey {
	for i := range app.Keys {
		if app.Keys[i].Type == KeyTypeSSH && app.Keys[i].Name == kname {
			return &app.Keys[i]
		}
	}
	return nil
}
