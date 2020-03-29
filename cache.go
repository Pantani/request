package request

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"github.com/Pantani/errors"
	"github.com/Pantani/logger"
	"github.com/patrickmn/go-cache"
	"net/url"
	"strings"
	"sync"
	"time"
)

// memCache memory cache object
type memCache struct {
	sync.RWMutex
	cache *cache.Cache
}

// deleteCache remove cache from memory
func (mc *memCache) deleteCache(key string) {
	mc.RLock()
	defer mc.RUnlock()
	mc.cache.Delete(key)
}

// setCache save cache inside memory with duration.
func (mc *memCache) setCache(key string, value interface{}, duration time.Duration) {
	mc.RLock()
	defer mc.RUnlock()
	b, err := json.Marshal(value)
	if err != nil {
		logger.Error(errors.E(err, "client cache cannot marshal cache object"))
		return
	}
	mc.cache.Set(key, b, duration)
}

// getCache restore cache from memory.
// It returns an error if occurs.
func (mc *memCache) getCache(key string, value interface{}) error {
	c, ok := mc.cache.Get(key)
	if !ok {
		return errors.E("validator cache: invalid cache key")
	}
	r, ok := c.([]byte)
	if !ok {
		return errors.E("validator cache: failed to cast cache to bytes")
	}
	err := json.Unmarshal(r, value)
	if err != nil {
		return errors.E(err, errors.Params{"key": key})
	}
	return nil
}

// generateKey generate a key to storage cache.
// It returns the key.
func (r *Request) generateKey(path string, query url.Values, body interface{}) string {
	var queryStr = ""
	if query != nil {
		queryStr = query.Encode()
	}
	requestUrl := strings.Join([]string{r.getUrl(path), queryStr}, "?")
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	hash := sha1.Sum(append([]byte(requestUrl), b...))
	return base64.URLEncoding.EncodeToString(hash[:])
}
