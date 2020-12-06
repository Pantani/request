package request

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Pantani/errors"
	"github.com/Pantani/logger"
	"github.com/patrickmn/go-cache"
)

var (
	_memoryCache *memCache
)

func init() {
	_memoryCache = &memCache{cache: cache.New(5*time.Minute, 5*time.Minute)}
}

// memCache represents the memory cache object.
type memCache struct {
	sync.RWMutex
	cache *cache.Cache
}

// PostWithCache sends an HTTP POST request with cache duration.
// After the duration the cache is expires and the request are made again.
// It returns an error if occurs.
//
// eg.:
// 	var block Block
//	err := c.PostWithCache(&block, "blocks/latest", CustomObject{Id: 3, Name: "request"}, time.Minute*20)
//
func (r *Request) PostWithCache(result interface{}, path string, body interface{}, cache time.Duration) error {
	key := r.generateKey(path, nil, body)
	err := _memoryCache.getCache(key, result)
	if err == nil {
		return nil
	}

	err = r.Post(result, path, body)
	if err != nil {
		return err
	}
	_memoryCache.setCache(key, result, cache)
	return err
}

// GetWithCache sends an HTTP GET request with cache duration.
// After the duration the cache is expires and the request are made again.
// It returns an error if occurs.
//
// eg.:
// 	var block Block
//	err := c.Get(&block, "blocks/latest", url.Values{"page": {"1"}}, time.Minute*20)
//
func (r *Request) GetWithCache(result interface{}, path string, query url.Values, cache time.Duration) error {
	key := r.generateKey(path, query, nil)
	err := _memoryCache.getCache(key, result)
	if err == nil {
		return nil
	}

	err = r.Get(result, path, query)
	if err != nil {
		return err
	}
	_memoryCache.setCache(key, result, cache)
	return err
}

// deleteCache delete cache from memory.
func (mc *memCache) deleteCache(key string) {
	mc.RLock()
	defer mc.RUnlock()
	_memoryCache.cache.Delete(key)
}

// setCache save the caches inside memory.
func (mc *memCache) setCache(key string, value interface{}, duration time.Duration) {
	mc.RLock()
	defer mc.RUnlock()
	b, err := json.Marshal(value)
	if err != nil {
		logger.Error(errors.E(err, "client cache cannot marshal cache object"))
		return
	}
	_memoryCache.cache.Set(key, b, duration)
}

// getCache returns the cache from memory. The response is unmarshal and
// stored inside the value parameter. It returns an error if occurs.
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
		return errors.E(err, "not found")
	}
	return nil
}

// generateKey generates a cache key from the request to save inside the memory.
func (r *Request) generateKey(path string, query url.Values, body interface{}) string {
	var queryStr = ""
	if query != nil {
		queryStr = query.Encode()
	}
	requestURL := strings.Join([]string{r.GetBase(path), queryStr}, "?")
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	hash := sha1.Sum(append([]byte(requestURL), b...))
	return base64.URLEncoding.EncodeToString(hash[:])
}
