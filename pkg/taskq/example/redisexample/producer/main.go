package main

import (
	"context"
	"flag"
	"log"

	"github.com/vmihailenco/taskq/example/redisexample"
)

func main() {
	flag.Parse()

	ctx := context.Background()
	go redisexample.LogStats()

	go func() {
		for {
			err := redisexample.MainQueue.AddJob(ctx, redisexample.CountTask.NewJob(ctx))
			if err != nil {
				log.Fatal(err)
			}
			redisexample.IncrLocalCounter()
		}
	}()

	sig := redisexample.WaitSignal()
	log.Println(sig.String())
}
