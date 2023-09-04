package changes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
)

var cacheTable = "GITWHO_CHANGES_CACHE"

func GetFromCache(opts ChangesOptions) (*ChangesResult, error) {
	// logrus.Debugf("Reusing results found in cache file")
	cachedb, err := utils.NewCacheDB(opts.CacheFile, cacheTable, opts.CacheTTLSeconds)
	if err != nil {
		return nil, fmt.Errorf("Cannot use file to cache results. err=%s", err)
	}
	defer cachedb.Close()

	cacheKey := getCacheKey(opts)
	cachedValue, err := cachedb.GetValue(cacheKey)
	if err != nil {
		return nil, err
	}
	if cachedValue == nil {
		logrus.Debugf("Cache miss for %s", cacheKey)
		return nil, nil
	}

	logrus.Debugf("Cache hit for %s", cacheKey)
	result := ChangesResult{}
	err = json.NewDecoder(strings.NewReader(*cachedValue)).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func SaveToCache(opts ChangesOptions, result ChangesResult) error {
	// logrus.Debugf("Saving results to cache file")

	cachedb, err := utils.NewCacheDB(opts.CacheFile, cacheTable, opts.CacheTTLSeconds)
	if err != nil {
		return err
	}
	defer cachedb.Close()

	cacheKey := getCacheKey(opts)

	b, err := json.Marshal(result)
	if err != nil {
		return err
	}
	resultStr := bytes.NewBuffer(b).String()

	err = cachedb.PutValue(cacheKey, resultStr)
	if err != nil {
		return err
	}
	logrus.Debugf("Cache saved for %s", cacheKey)

	return nil
}

func getCacheKey(opts ChangesOptions) string {
	// since/until might be "relative" to current time ("1 day ago" etc)
	// so in cases those fields are not absolute dates, use current date day
	// as part of the key
	// the effect is that this cache will be valid only for one day
	add := ""
	_, errSince := time.Parse(time.DateOnly, opts.Since)
	_, errUntil := time.Parse(time.DateOnly, opts.Until)
	if errSince != nil || errUntil != nil {
		add = time.Now().Format(time.DateOnly)
	}

	return fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%s:%s",
		opts.RepoDir,
		opts.Branch,
		opts.AuthorsRegex,
		opts.AuthorsNotRegex,
		opts.FilesRegex,
		opts.FilesNotRegex,
		opts.Since,
		opts.Until,
		add)
}
