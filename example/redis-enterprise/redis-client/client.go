package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	tr = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient = &http.Client{Transport: tr}
)

func main() {
	ctx := context.Background()

	if err := createDatabase(ctx); err != nil {
		log.Println(err)
	}

	go func() {
		rdb := redis.NewClient(&redis.Options{
			Addr: ":12000",
		})
		genLoad(ctx, rdb)
	}()

	go func() {
		rdb := redis.NewClient(&redis.Options{
			Addr: ":12001",
		})
		genLoad(ctx, rdb)
	}()

	select {}
}

func createDatabase(ctx context.Context) error {
	b, err := json.Marshal(map[string]any{
		"name":         "mydb",
		"memory_size":  10 << 20,
		"type":         "redis",
		"port":         12000,
		"sharding":     true,
		"shards_count": 2,
	})
	if err != nil {
		return err
	}

	endpoint := "https://localhost:9444/v1/bdbs"
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.SetBasicAuth("vladimir.webdev@gmail.com", "test")
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("%s: %.100q", resp.Status, body)
	}

	return nil
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
