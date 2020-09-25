package caches

import (
	"time"
)

// TimeLocationCache is the default loading cache used for caching *time.Location
// object If the default size is too big/small and/or a cache limit isn't desired
// at all, caller can simply replace the cache during global initialization. But
// be aware it's global so any packages uses this package inside your process will
// be affected.
var TimeLocationCache = NewLoadingCache()

func GetTimeLocation(tz string) (*time.Location, error) {
	loc, err := TimeLocationCache.Get(tz, func(key interface{}) (interface{}, error) {
		return time.LoadLocation(key.(string))
	})
	if err != nil {
		return nil, err
	}
	return loc.(*time.Location), nil
}
