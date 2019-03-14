package cache

import (
	"encoding/json"
	"sync"
	"time"
)

type Cache interface {
	Get(key string) (string, bool)
	GetBytes(key string) []byte
	GetInterface(key string, resp interface{}) error
	Set(key, val string, ttl time.Duration)
	SetBytes(key string, val []byte, ttl time.Duration)
	SetInterface(key string, val interface{}, ttl time.Duration) error
	TTL(key string) time.Duration
	Expire(key string, ttl time.Duration) bool
}

type cacheImpl struct {
	vals map[string][]byte
	ttls map[string]time.Time
	mu   sync.Mutex
}

func New() Cache {
	return &cacheImpl{
		vals: make(map[string][]byte),
		ttls: make(map[string]time.Time),
	}
}

func (r *cacheImpl) GetBytes(key string) []byte {
	r.mu.Lock()
	defer r.mu.Unlock()

	val, ok := r.vals[key]
	if !ok {
		return nil
	}
	ttl, ok := r.ttls[key]
	if !ok {
		return nil
	}
	if ttl.Before(time.Now()) {
		delete(r.ttls, key)
		delete(r.vals, key)
		return nil
	}

	return val
}

func (r *cacheImpl) SetBytes(key string, val []byte, ttl time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.vals[key] = val
	r.ttls[key] = time.Now().Add(ttl)
}

func (r *cacheImpl) Get(key string) (string, bool) {
	bs := r.GetBytes(key)
	if bs == nil {
		return "", false
	}
	return string(bs), true
}

func (r *cacheImpl) Set(key, val string, ttl time.Duration) {
	r.SetBytes(key, []byte(val), ttl)
}

func (r *cacheImpl) GetInterface(key string, resp interface{}) error {
	bs := r.GetBytes(key)
	if bs != nil {
		return json.Unmarshal(bs, &resp)
	}
	return nil
}

func (r *cacheImpl) SetInterface(key string, val interface{}, ttl time.Duration) error {
	bs, err := json.Marshal(val)
	if err != nil {
		return err
	}
	r.SetBytes(key, bs, ttl)
	return nil
}

func (r *cacheImpl) TTL(key string) time.Duration {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.vals[key]; !ok {
		return -time.Second
	}

	ttl2, ok := r.ttls[key]
	if !ok {
		return -time.Second
	}

	now := time.Now()
	if ttl2.Before(now) {
		delete(r.ttls, key)
		delete(r.vals, key)
		return -time.Second
	}

	return ttl2.Sub(now)
}

func (r *cacheImpl) Expire(key string, ttl time.Duration) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.vals[key]; !ok {
		return false
	}

	if _, ok := r.ttls[key]; !ok {
		return false
	}

	r.ttls[key] = time.Now().Add(ttl)

	return true
}

func NearlyEqual(t1, t2 time.Duration) bool {
	if t1 < t2 {
		t1, t2 = t2, t1
	}
	// t1 >= t2
	return t1 >= t2 && t1-time.Millisecond <= t2
}
