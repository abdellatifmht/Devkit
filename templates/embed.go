// Package templates embeds the templates shipped with the binary.
// A template is a directory containing manifest.yaml, base/, and modules/.
package templates

import "embed"

//go:embed all:rag-python all:agent-python all:mlops-python all:fastapi-python all:nextjs all:spring-boot-react
var FS embed.FS
