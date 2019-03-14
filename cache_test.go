package cache_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Chyroc/memcache"
)

func TestCacheImpl(t *testing.T) {
	as := assert.New(t)
	c := cache.New()

	// exist
	{
		// set
		c.Set("k1", "v1", time.Second*2)

		// get
		v1, ok := c.Get("k1")
		as.True(ok)
		as.Equal("v1", v1)

		// ttl
		as.Equal(true, time.Second*2-time.Millisecond <= c.TTL("k1") && c.TTL("k1") <= time.Second*2)

		// expire
		c.Expire("k1", time.Minute)
		as.Equal(true, time.Minute-time.Millisecond <= c.TTL("k1") && c.TTL("k1") <= time.Minute)
	}

	// not-exist
	{
		v2, ok := c.Get("k2")
		as.False(ok)
		as.Equal("", v2)

		as.Equal(-time.Second, c.TTL("v2"))
	}

	// expire
	{
		c.Set("k3", "v3", time.Second)

		time.Sleep(time.Second)

		v3, ok := c.Get("k3")
		as.False(ok)
		as.Equal("", v3)

		as.Equal(-time.Second, c.TTL("k3"))
	}
}
