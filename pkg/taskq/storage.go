package taskq

import (
	"context"
	"sync"
	"time"

	"github.com/hashicorp/golang-lru/simplelru"
)

type Storage interface {
	Exists(ctx context.Context, key string) bool
}

var _ Storage = (*localStorage)(nil)
var _ Storage = (*redisStorage)(nil)

// LOCAL

type localStorage struct {
	mu    sync.Mutex
	cache *simplelru.LRU
}

func NewLocalStorage() Storage {
	return &localStorage{}
}

func (s *localStorage) Exists(_ context.Context, key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cache == nil {
		var err error
		s.cache, err = simplelru.NewLRU(128000, nil)
		if err != nil {
			panic(err)
		}
	}

	_, ok := s.cache.Get(key)
	if ok {
		return true
	}

	s.cache.Add(key, nil)
	return false
}

// REDIS

type redisStorage struct {
	redis Redis
}

func newRedisStorage(redis Redis) Storage {
	return &redisStorage{
		redis: redis,
	}
}

func (s *redisStorage) Exists(ctx context.Context, key string) bool {
	val, err := s.redis.SetNX(ctx, key, "", 24*time.Hour).Result()
	if err != nil {
		return true
	}
	return !val
}
