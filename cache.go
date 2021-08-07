package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	strCache = map[string]bool{}
	prefix   = "bingimage"
	p        = false // p for persistent
	r        = redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})
	ctx = context.Background()
)

func toKey(s string) string {
	return fmt.Sprintf("%v-%v-%v", prefix, time.Now().Format("202108"), s)
}

func keySeen(key string) (seen bool, err error) {
	_, err = r.Get(ctx, key).Result()
	if err == nil {
		return true, nil
	}
	if err == redis.Nil {
		return false, nil
	}
	return false, err
}

func strSeen(path string) (seen bool, err error) {
	seen = false
	key := toKey(path)

	if _, found := strCache[key]; found {
		seen = true
		return
	}

	if p {
		seen, err = keySeen(key)
	}

	return
}

func setSeen(strs ...string) {
	for _, s := range strs {
		key := toKey(s)
		strCache[key] = true

		if p {
			r.Set(ctx, key, true, 0) // do not expire
		}
	}
}
