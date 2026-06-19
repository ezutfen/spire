package spa

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	frontenddist "github.com/EQEmu/spire/frontend"
	"github.com/EQEmu/spire/internal/logger"

	"github.com/EQEmu/spire/internal/env"
)

type Spa struct {
	logger *logger.AppLogger
	spa    *Packer
}

func (s Spa) Spa() *Packer {
	return s.spa
}

// Spire SPA vars
const (
	SpireBasePath      = ""
	SpireLocalBasePath = "../../../frontend/dist/"
	SpireSpaIndex      = "index.html"
)

func NewSpa(logger *logger.AppLogger) *Spa {
	assets, err := fs.Sub(frontenddist.DistFS, "dist")
	if err != nil {
		logger.Error().Err(err).Msg("error creating embedded frontend fs, falling back to local dist")
		assets = os.DirFS(filepath.Clean(SpireLocalBasePath))
	}

	return &Spa{
		logger: logger,
		spa: NewPackedSpaService(
			logger,
			assets,
			PackedSpaServeConfig{
				BasePath:      SpireBasePath,
				LocalBasePath: SpireLocalBasePath,
				SpaIndex:      SpireSpaIndex,
				SkipPaths:     strings.Split(env.Get("SPA_SKIP_PATH_PREFIXES", "/auth,/api,/swagger,/websocket,/eqsage,/static"), ","),
			},
		),
	}
}
