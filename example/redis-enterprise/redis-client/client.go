package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
)

func main() {
	ctx := context.Background()

	go func() {
		rdb := redis.NewClient(&redis.Options{
			Addr: "redis1:12000",
		})
		genLoad(ctx, rdb)
	}()

	go func() {
		rdb := redis.NewClient(&redis.Options{
			Addr: "redis2:12000",
		})
		genLoad(ctx, rdb)
	}()

	select {}
}

func genLoad(ctx context.Context, rdb *redis.Client) {
	for {
		if err := rdb.Set(ctx, fmt.Sprint(rand.Intn(100)), rand.Intn(1000), 0).Err(); err != nil {
			log.Println(err)
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if err := rdb.Get(ctx, fmt.Sprint(rand.Intn(1000))).Err(); err != nil && err != redis.Nil {
			log.Println(err)
			time.Sleep(500 * time.Millisecond)
			continue
		}
	}
}
