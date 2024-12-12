// Package run implements an actor-runner with deterministic teardown. It is
// somewhat similar to package errgroup, except it does not require actor
// goroutines to understand context semantics. This makes it suitable for use in
// more circumstances; for example, goroutines which are handling connections
// from net.Listeners, or scanning input from a closable io.Reader.
package run

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type StartFunc func() error
type OnStopFunc func(ctx context.Context, err error) error

// Group collects actors (functions) and runs them concurrently.
// When one actor (function) returns, all actors are interrupted.
// The zero value of a Group is usable.
type Group struct {
	execute   []func() error
	onStop    hooks
	onStopped hooks

	stopTimeout time.Duration
}

func NewGroup() *Group {
	return &Group{
		stopTimeout: 30 * time.Second,
	}
}

// Add an actor (function) to the group. Each actor must be pre-emptable by an
// interrupt function. That is, if interrupt is invoked, execute should return.
// Also, it must be safe to call interrupt even after execute has returned.
//
// The first actor (function) to return interrupts all running actors.
// The error is passed to the interrupt functions, and is returned by Run.
func (g *Group) Add(name string, fn StartFunc) {
	g.execute = append(g.execute, func() error {
		if err := fn(); err != nil {
			return fmt.Errorf("%s failed: %w", name, err)
		}
		return nil
	})
}

func (g *Group) OnStop(fn OnStopFunc) {
	g.onStop.Add(fn)
}

func (g *Group) OnStopped(fn OnStopFunc) {
	g.onStopped.Add(fn)
}

// Run all actors (functions) concurrently.
// When the first actor returns, all others are interrupted.
// Run only returns when all actors have exited.
// Run returns the error returned by the first exiting actor.
func (g *Group) Run() error {
	errors := make(chan error, len(g.execute))
	var err error

	if len(g.execute) > 0 {
		// Run each actor.
		for _, fn := range g.execute {
			fn := fn
			go func() {
				errors <- fn()
			}()
		}

		// Wait for the first actor to stop.
		err = <-errors
	}

	ctx, cancel := context.WithTimeout(context.Background(), g.stopTimeout)
	defer cancel()

	if err := g.onStop.Run(ctx, err); err != nil {
		return err
	}
	if err := g.onStopped.Run(ctx, err); err != nil {
		return err
	}

	done := make(chan struct{})
	go func() {
		// Wait for all actors to stop.
		for i := 1; i < cap(errors); i++ {
			<-errors
		}
		close(done)
	}()

	select {
	case <-done:
		// Return the original error.
		return err
	case <-ctx.Done():
		slog.Error("can't stop the run group in time", slog.Any("err", err))
		// Return the original error.
		return err
	}
}

func (g *Group) WaitExitSignal() {
	g.WaitSignal(syscall.SIGINT, syscall.SIGTERM)
}

func (g *Group) WaitSignal(signals ...os.Signal) {
	ctx, cancel := context.WithCancel(context.Background())
	g.Add("wait-signal", func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, signals...)
		defer signal.Stop(c)
		select {
		case sig := <-c:
			return SignalError{Signal: sig}
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	g.OnStop(func(context.Context, error) error {
		cancel()
		return nil
	})
}

// SignalError is returned by the signal handler's execute function
// when it terminates due to a received signal.
type SignalError struct {
	Signal os.Signal
}

// Error implements the error interface.
func (e SignalError) Error() string {
	return fmt.Sprintf("received signal %s", e.Signal)
}

//------------------------------------------------------------------------------

type hooks struct {
	fns []OnStopFunc
}

func (hs *hooks) Add(fn OnStopFunc) {
	hs.fns = append(hs.fns, fn)
}

func (hs hooks) Run(ctx context.Context, err error) error {
	var wg sync.WaitGroup

	for _, fn := range hs.fns {
		fn := fn
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := fn(ctx, err); err != nil {
				fmt.Println(err)
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
