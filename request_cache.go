package blockatlas

import (
	"github.com/patrickmn/go-cache"
	"net/url"
	"time"
)

var (
	memoryCache *memCache
)

func init() {
	memoryCache = &memCache{cache: cache.New(5*time.Minute, 5*time.Minute)}
}

// Get make a GET request with query url values and cache.
// It returns an error if occurs.
func (r *Request) GetWithCache(result interface{}, path string, query url.Values, cache time.Duration) error {
	key := r.generateKey(path, query, nil)
	err := memoryCache.getCache(key, result)
	if err == nil {
		return nil
	}

	err = r.Get(result, path, query)
	if err != nil {
		return err
	}
	memoryCache.setCache(key, result, cache)
	return err
}

// Get make a POST request with body value and cache.
// It returns an error if occurs.
func (r *Request) PostWithCache(result interface{}, path string, body interface{}, cache time.Duration) error {
	key := r.generateKey(path, nil, body)
	err := memoryCache.getCache(key, result)
	if err == nil {
		return nil
	}

	err = r.Post(result, path, body)
	if err != nil {
		return err
	}
	memoryCache.setCache(key, result, cache)
	return err
}
