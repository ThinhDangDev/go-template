package generator

import (
	"strings"
	"text/template"
	"unicode"
)

// FuncMap returns custom template functions
func FuncMap() template.FuncMap {
	return template.FuncMap{
		// String transformations
		"lower":     strings.ToLower,
		"upper":     strings.ToUpper,
		"title":     strings.Title,
		"snake":     toSnakeCase,
		"camel":     toCamelCase,
		"pascal":    toPascalCase,
		"kebab":     toKebabCase,
		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,

		// API type checks
		"hasREST": func(apiType string) bool {
			return apiType == "rest" || apiType == "both"
		},
		"hasGRPC": func(apiType string) bool {
			return apiType == "grpc" || apiType == "both"
		},

		// Auth type checks
		"hasJWT": func(authType string) bool {
			return authType == "jwt" || authType == "both"
		},
		"hasOAuth2": func(authType string) bool {
			return authType == "oauth2" || authType == "both"
		},
		"hasAuth": func(authType string) bool {
			return authType != "none"
		},
	}
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func toCamelCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for i := range parts {
		if i == 0 {
			parts[i] = strings.ToLower(parts[i])
		} else {
			parts[i] = strings.Title(parts[i])
		}
	}
	return strings.Join(parts, "")
}

func toPascalCase(s string) string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

func toKebabCase(s string) string {
	return strings.ReplaceAll(toSnakeCase(s), "_", "-")
}
