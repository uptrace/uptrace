/*
This package contains an example usage of redis-backed taskq. It includes two programs:

1. api that sends messages on the queue
2. worker that receives messages

They share common definitions for the queue `MainQueue` with a single task defined on it `CountTask`. api runs in a loop
submitting messages to the queue, for each message it submits it increments a process-local counter. worker starts a
consumer that executes the `CountTask` for each message receives - the task handler also implements a process local
counter. If you run each program in a separate terminal with:

   go run ./examples/redisexample/api/main.go
   go run ./examples/redisexample/worker/main.go

Then api will periodically print the messages sent to the terminal and worker will print the message received.

If you spawn further workers they will share consumption of the messages produced by api.
*/
package redisexample
