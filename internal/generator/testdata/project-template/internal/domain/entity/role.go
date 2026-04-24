package entity

import "strings"

const (
	RoleAdmin    = "admin"
	RoleOperator = "operator"
	RoleViewer   = "viewer"
)

type RoleDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NormalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func IsValidRole(role string) bool {
	switch NormalizeRole(role) {
	case RoleAdmin, RoleOperator, RoleViewer:
		return true
	default:
		return false
	}
}

func AvailableRoles() []RoleDefinition {
	return []RoleDefinition{
		{
			Name:        RoleAdmin,
			Description: "Full access to all APIs and user access management.",
		},
		{
			Name:        RoleOperator,
			Description: "Authenticated non-admin access for internal operators.",
		},
		{
			Name:        RoleViewer,
			Description: "Default authenticated access for self-service users.",
		},
	}
}
