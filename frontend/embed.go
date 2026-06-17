package frontenddist

import "embed"

var (
	// DistFS embeds the Vite production bundle consumed by the Go HTTP server.
	//
	//go:embed all:dist
	DistFS embed.FS
)
