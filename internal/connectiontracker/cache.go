package connectiontracker

import (
	"context"
	"fmt"
	"github.com/dgraph-io/ristretto"
	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

type connCache struct {
	// cache.Cache is a thread-safe implementation of a hashmap with a TinyLFU admission
	// policy and a Sampled LFU eviction policy. You can use the same Cache instance
	// from as many goroutines as you want.
	manager  *cache.Cache[*ConnEntry]
	cacheTTL time.Duration
	mutex    sync.RWMutex
}

// newCacheManager creates the instance of Cache, currently using gocache + ristretto
func newCacheManager(ttl time.Duration) *connCache {
	// TODO: spend some time thinking about cache config
	// TODO: add prometheus Metrics
	ristrettoCache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1000,
		MaxCost:     100,
		BufferItems: 64,
	})
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	ristrettoStore := store.NewRistretto(ristrettoCache)
	cacheManager := cache.New[*ConnEntry](ristrettoStore)
	return &connCache{
		manager:  cacheManager,
		cacheTTL: ttl,
	}
}

func (c *connCache) getOrSet(ctx context.Context, conn *ConnEntry) *ConnEntry {
	key := fmt.Sprintf("%s->%s", conn.SrcIP, conn.DstIP)
	found, err := c.manager.Get(ctx, key)
	if err != nil {
		err := c.manager.Set(ctx, key, conn, store.WithExpiration(c.cacheTTL))
		if err != nil {
			log.Err(err).Send()
		}
		return conn
	}
	// TODO: Mutex is only needed for this part, updating values in a map
	// I need concurrent safe Set structure instead of map[int]bool
	// Doesn't really matter because new goroutine would override same value
	c.mutex.Lock()
	ports := updatePorts(conn.Ports, found.Ports)
	found.Ports = ports
	c.mutex.Unlock()
	errSet := c.manager.Set(ctx, key, found, store.WithExpiration(c.cacheTTL))
	if errSet != nil {
		log.Err(errSet)
	}
	return found
}

func updatePorts(connPorts, existingPorts map[int]bool) map[int]bool {
	ports := existingPorts
	for k, v := range connPorts {
		ports[k] = v
	}
	return ports
}
