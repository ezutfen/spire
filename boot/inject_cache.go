package boot

import (
	gocache "github.com/patrickmn/go-cache"
	"time"
)

// provides cache
func provideCache() *gocache.Cache {
	return gocache.New(5*time.Minute, 10*time.Minute)
}
