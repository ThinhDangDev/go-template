package prompt

import (
	"fmt"
	"regexp"

	"github.com/AlecAivazis/survey/v2"
	"github.com/thinhdang/go-template/internal/config"
)

var modulePathRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/-]*$`)

// CollectConfig prompts user for project configuration
func CollectConfig(projectName string) (*config.ProjectConfig, error) {
	cfg := config.NewDefaultConfig(projectName)

	questions := []*survey.Question{
		{
			Name: "modulePath",
			Prompt: &survey.Input{
				Message: "Go module path:",
				Default: cfg.ModulePath,
			},
			Validate: func(val interface{}) error {
				str := val.(string)
				if !modulePathRegex.MatchString(str) {
					return fmt.Errorf("invalid module path")
				}
				return nil
			},
		},
		{
			Name: "description",
			Prompt: &survey.Input{
				Message: "Project description:",
				Default: cfg.Description,
			},
		},
		{
			Name: "apiType",
			Prompt: &survey.Select{
				Message: "API type:",
				Options: []string{"REST (Gin)", "gRPC", "Both"},
				Default: "REST (Gin)",
			},
		},
		{
			Name: "authType",
			Prompt: &survey.Select{
				Message: "Authentication:",
				Options: []string{"JWT", "OAuth2", "Both", "None"},
				Default: "JWT",
			},
		},
		{
			Name: "withDocker",
			Prompt: &survey.Confirm{
				Message: "Generate Docker files?",
				Default: true,
			},
		},
		{
			Name: "ciProvider",
			Prompt: &survey.Select{
				Message: "CI/CD provider:",
				Options: []string{"GitHub Actions", "GitLab CI", "None"},
				Default: "GitHub Actions",
			},
		},
		{
			Name: "withMonitor",
			Prompt: &survey.Confirm{
				Message: "Include Prometheus + Grafana monitoring?",
				Default: true,
			},
		},
	}

	answers := struct {
		ModulePath  string
		Description string
		APIType     string
		AuthType    string
		WithDocker  bool
		CIProvider  string
		WithMonitor bool
	}{}

	if err := survey.Ask(questions, &answers); err != nil {
		return nil, err
	}

	// Map answers to config
	cfg.ModulePath = answers.ModulePath
	cfg.Description = answers.Description
	cfg.WithDocker = answers.WithDocker
	cfg.WithMonitor = answers.WithMonitor

	// Map API type
	switch answers.APIType {
	case "REST (Gin)":
		cfg.APIType = config.APITypeREST
	case "gRPC":
		cfg.APIType = config.APITypeGRPC
	case "Both":
		cfg.APIType = config.APITypeBoth
	}

	// Map auth type
	switch answers.AuthType {
	case "JWT":
		cfg.AuthType = config.AuthTypeJWT
	case "OAuth2":
		cfg.AuthType = config.AuthTypeOAuth2
	case "Both":
		cfg.AuthType = config.AuthTypeBoth
	case "None":
		cfg.AuthType = config.AuthTypeNone
	}

	// Map CI provider
	switch answers.CIProvider {
	case "GitHub Actions":
		cfg.WithCI = "github"
	case "GitLab CI":
		cfg.WithCI = "gitlab"
	case "None":
		cfg.WithCI = "none"
	}

	return cfg, nil
}
