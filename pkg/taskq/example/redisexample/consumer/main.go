package main

import (
	"context"
	"flag"
	"log"

	"github.com/vmihailenco/taskq/example/redisexample"
)

func main() {
	flag.Parse()

	c := context.Background()

	err := redisexample.QueueFactory.StartConsumers(c)
	if err != nil {
		log.Fatal(err)
	}

	go redisexample.LogStats()

	sig := redisexample.WaitSignal()
	log.Println(sig.String())

	err = redisexample.QueueFactory.Close()
	if err != nil {
		log.Fatal(err)
	}
}
