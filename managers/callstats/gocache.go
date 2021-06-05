package callstats

import "github.com/patrickmn/go-cache"

func SetLocalCache(key string, value interface{})  {
	appCache.Set(key, value, cache.DefaultExpiration)
}

func GetLocalCache(key string)  interface{}{
	value, found := appCache.Get(key)
	if found {
		return value
	}
	return nil
}