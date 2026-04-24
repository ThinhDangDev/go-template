package auth

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

type Authorizer struct {
	enforcer *casbin.SyncedCachedEnforcer
}

func NewAuthorizer(db *sql.DB, modelPath string) (*Authorizer, error) {
	m, err := model.NewModelFromFile(modelPath)
	if err != nil {
		return nil, fmt.Errorf("load casbin model: %w", err)
	}

	enforcer, err := casbin.NewSyncedCachedEnforcer(m)
	if err != nil {
		return nil, fmt.Errorf("create casbin enforcer: %w", err)
	}

	if err := loadPolicies(db, enforcer); err != nil {
		return nil, fmt.Errorf("load casbin policies: %w", err)
	}

	return &Authorizer{enforcer: enforcer}, nil
}

func (a *Authorizer) Authorize(role, object, action string) (bool, error) {
	return a.enforcer.Enforce(role, object, action)
}

func (a *Authorizer) Reload(db *sql.DB) error {
	a.enforcer.ClearPolicy()
	return loadPolicies(db, a.enforcer)
}

func loadPolicies(db *sql.DB, enforcer *casbin.SyncedCachedEnforcer) error {
	rows, err := db.Query(`
		SELECT ptype, v0, v1, v2, v3, v4, v5
		FROM casbin_rule
		ORDER BY id ASC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var ptype, v0, v1, v2, v3, v4, v5 string
		if err := rows.Scan(&ptype, &v0, &v1, &v2, &v3, &v4, &v5); err != nil {
			return err
		}

		rule := trimTrailingEmpty(v0, v1, v2, v3, v4, v5)
		if len(rule) == 0 {
			continue
		}
		ruleArgs := toInterfaceSlice(rule)

		if strings.HasPrefix(ptype, "g") {
			if _, err := enforcer.AddNamedGroupingPolicy(ptype, ruleArgs...); err != nil {
				return err
			}
			continue
		}

		if _, err := enforcer.AddNamedPolicy(ptype, ruleArgs...); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return enforcer.BuildRoleLinks()
}

func trimTrailingEmpty(values ...string) []string {
	last := len(values)
	for last > 0 && strings.TrimSpace(values[last-1]) == "" {
		last--
	}

	out := make([]string, 0, last)
	for i := 0; i < last; i++ {
		out = append(out, values[i])
	}

	return out
}

func toInterfaceSlice(values []string) []interface{} {
	out := make([]interface{}, 0, len(values))
	for _, value := range values {
		out = append(out, value)
	}

	return out
}
