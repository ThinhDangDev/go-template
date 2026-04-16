package templates

import "embed"

//go:embed all:base all:clean-arch all:docker all:monitoring all:ci-cd all:tests
var FS embed.FS
