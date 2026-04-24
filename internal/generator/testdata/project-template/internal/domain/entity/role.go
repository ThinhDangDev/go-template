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
			Description: "Operational access for protected operator routes.",
		},
		{
			Name:        RoleViewer,
			Description: "Read-only access for protected viewer routes.",
		},
	}
}
