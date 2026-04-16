package config

// APIType represents the API delivery type
type APIType string

const (
	APITypeREST APIType = "rest"
	APITypeGRPC APIType = "grpc"
	APITypeBoth APIType = "both"
)

// AuthType represents authentication method
type AuthType string

const (
	AuthTypeJWT    AuthType = "jwt"
	AuthTypeOAuth2 AuthType = "oauth2"
	AuthTypeBoth   AuthType = "both"
	AuthTypeNone   AuthType = "none"
)

// ProjectConfig holds all project configuration
type ProjectConfig struct {
	Name        string   // Project name (directory name)
	ModulePath  string   // Go module path
	Description string   // Project description
	APIType     APIType  // REST, gRPC, or both
	AuthType    AuthType // JWT, OAuth2, both, or none
	Database    string   // Database type (postgres default)
	WithDocker  bool     // Generate Docker files
	WithCI      string   // CI provider (github, gitlab, none)
	WithMonitor bool     // Generate monitoring configs
}

// NewDefaultConfig creates config with sensible defaults
func NewDefaultConfig(name string) *ProjectConfig {
	return &ProjectConfig{
		Name:        name,
		ModulePath:  "github.com/user/" + name,
		Description: "A Go backend service",
		APIType:     APITypeREST,
		AuthType:    AuthTypeJWT,
		Database:    "postgres",
		WithDocker:  true,
		WithCI:      "github",
		WithMonitor: true,
	}
}
