package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSaveGetCache(t *testing.T) {
	// os.Remove("gitwho.cache")

	cachedb, err := NewCacheDB("gitwho.cache", "TEST_CACHE", 1)
	require.Nil(t, err)

	err = cachedb.PutValue("key", "value")
	require.Nil(t, err)

	// contents not expired
	value, err := cachedb.GetValue("key")
	require.Nil(t, err)
	require.NotNil(t, value)
	require.Equal(t, "value", *value)

	time.Sleep(1100 * time.Millisecond)

	// contents expired
	value, err = cachedb.GetValue("key")
	require.Nil(t, err)
	require.Nil(t, value)

}
