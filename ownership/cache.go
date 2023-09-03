package ownership

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
)

var cacheTable = "GITWHO_CACHE"

func GetFromCache(opts OwnershipOptions) (*OwnershipResult, error) {
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
	result := OwnershipResult{}
	err = json.NewDecoder(strings.NewReader(*cachedValue)).Decode(&result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func SaveToCache(opts OwnershipOptions, result OwnershipResult) error {
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

func getCacheKey(opts OwnershipOptions) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s:%s:%s:%d",
		opts.RepoDir,
		opts.CommitId,
		opts.Branch,
		opts.AuthorsRegex,
		opts.AuthorsNotRegex,
		opts.FilesRegex,
		opts.FilesNotRegex,
		opts.MinDuplicateLines)
}
