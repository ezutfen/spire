package spa

import (
	"github.com/EQEmu/spire/internal/http/routes"
	"github.com/EQEmu/spire/internal/logger"
	"github.com/labstack/echo/v4"
	"io/fs"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// Packaged SPA is a single page application that is compiled into the resulting
// go application binary and ultimately served using echo in our monorepo framework
//
// The production SPA is embedded into the Go binary and served by echo.

type Packer struct {
	fileServer http.Handler
	handler    echo.HandlerFunc
	assets     fs.FS
	logger     *logger.AppLogger
	config     PackedSpaServeConfig
}

func (s Packer) Handler() echo.HandlerFunc {
	return s.handler
}

// PackedSpaServeConfig is the configuration for the PackedSpaService
type PackedSpaServeConfig struct {
	BasePath      string   // Root path from where the SPA is served
	LocalBasePath string   // The path that the SPA is located from the relative context of this provider
	SpaIndex      string   // SPA index - where requests will fallback to
	SkipPaths     []string // Route prefixes to skip entirely
}

// NewPackedSpaService creates a new instance of the PackedSpaService
func NewPackedSpaService(logger *logger.AppLogger, assets fs.FS, config PackedSpaServeConfig) *Packer {
	s := &Packer{}
	s.config = config
	s.logger = logger
	s.assets = assets
	s.fileServer = http.FileServer(http.FS(s.assets))
	s.handler = WrapCachedHandler(http.StripPrefix(s.config.BasePath, s.fileServer))

	return s
}

func WrapCachedHandler(h http.Handler) echo.HandlerFunc {
	return func(c echo.Context) error {
		if contains([]string{".js", ".css", ".png", ".woff", ".ttf", ".jpg", ".gif", ".svg", ".ico"}, c.Request().RequestURI) {
			c.Response().Header().Set("Vary", "Accept-Encoding")
			c.Response().Header().Set("Cache-Control", "public, max-age=7776000")
		}
		h.ServeHTTP(c.Response(), c.Request())
		return nil
	}
}

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func contains(slice []string, val string) bool {
	for _, item := range slice {
		if strings.Contains(val, item) {
			return true
		}
	}
	return false
}

// MiddlewareHandler returns a middleware handler for the PackedSpaService
// This middleware handler is for the most part a static middleware handler for handling static assets
// The main difference between this middleware and a generic static middleware is that it provides assets that live
// within a static asset compiler that gets bundled within the compiled resulting binary /packed-spa/index.html
func (s Packer) MiddlewareHandler() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			// if we explicitly ignore paths, process routing as normal
			for _, skipPath := range s.config.SkipPaths {
				if strings.Contains(c.Request().URL.Path, skipPath) {
					return next(c)
				}
			}

			p := c.Request().URL.Path
			if strings.HasSuffix(c.Path(), "*") { // When serving from a group, e.g. `/static*`.
				p = c.Param("*")
			}
			p, err = url.PathUnescape(p)
			if err != nil {
				return
			}

			requestPath := path.Clean(strings.TrimPrefix(p, "/"))
			if requestPath == "." {
				index, err := fs.ReadFile(s.assets, s.config.SpaIndex)
				if err != nil {
					s.logger.Error().Err(err).Msg("error finding spa index")
				}
				return c.HTML(http.StatusOK, string(index))
			}

			// If we find a valid non-index file in the box, continue the request as normal
			// and let the static asset handler pick up the request later
			fileRequest := path.Clean(strings.TrimPrefix(strings.Replace(c.Request().URL.Path, s.config.BasePath, "", 1), "/"))
			_, err = fs.Stat(s.assets, fileRequest)
			if err == nil {
				return next(c)
			}

			// If we didn't find a non-index asset at this point, we need to return the
			// spa index when nested SPA route requests are made
			// eg: /spa/nested/route
			index, err := fs.ReadFile(s.assets, s.config.SpaIndex)
			if err != nil {
				s.logger.Error().Err(err).Msg("error finding spa index")
				return next(c)
			}

			return c.HTML(http.StatusOK, string(index))
		}
	}
}

// PackagedSpaController is a provider to easily handle registration of the necessary route handlers
type PackagedSpaController struct {
	service *Packer
}

func NewSpaController(s *Packer) *PackagedSpaController {
	return &PackagedSpaController{
		service: s,
	}
}

// Routes registers controller specific routes
func (c *PackagedSpaController) Routes() []*routes.Route {
	return []*routes.Route{
		routes.RegisterRoute(http.MethodGet, "/*", c.service.Handler(), nil),
	}
}

func (s Packer) Controller() *PackagedSpaController {
	return NewSpaController(&s)
}
