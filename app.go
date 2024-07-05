// Package app provides light-weight utilities around app life cycle events.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

//go:generate go-enumerator
type state uint

const (
	invalid state = iota
	running
	stopping
	terminated
)

type actionName string

const (
	runMainJob         actionName = "RunMainJob"
	registerCleanupJob actionName = "RegisterCleanupJob"
)

// At its core, and app is a place to register work to be
// done during the life-cycle of the app. And, a central
// place to signal when to begin shutdown. The zero value
// of an App is invalid; use [New] to get a new instance
// of an App.
//
// Once an app is initialized by calling [New], it can
// be in one of three states.
//
// Running - the normal running state. In this state, it
// is valid to queue more work by calling [App.Go] and to
// queue more shutdown jobs by calling [App.OnShutdown].
// An app will transition out of the running state once
// any of its tasks, queued via [App.Go], return a non-nil
// error, or if its parent context is cancelled.
//
// Stopping - a limited period of time after shutdown has
// been signaled to allow for clean-up and finalization
// tasks to run. In this state, it is valid to queue more
// work by calling [App.OnShutdown]. It is invalid to call
// [App.Go] in this phase.
//
// Terminated - after all clean-up tasks have complete,
// or the shutdown grace period has elapsed, the app is
// in its final terminal state. New work can no longer be
// queued.
type App struct {
	keepAlive bool

	lock  sync.RWMutex
	state state

	ctx    context.Context
	cancel context.CancelCauseFunc

	grp    *errgroup.Group
	grpCtx context.Context

	shutdownQ   []func() error
	shutdownCtx context.Context
	shutdownGrp *errgroup.Group

	shutdownTimeout time.Duration
	log             *slog.Logger
}

const (
	DefaultShutdownTimeout = time.Second * 15
)

// New initializes an new app for use. By default,
// it will have the [DefaultShutdownTimeout], it
// will use [slog.Default] to log, and it will
// be set to keep-alive. These options can be modified
// by passing the appropriate opts.
func New(parent context.Context, opts ...AppOptFunc) *App {
	ctx, cancel := context.WithCancelCause(parent)
	grp, grpCtx := errgroup.WithContext(ctx)

	ret := &App{
		state:           running,
		ctx:             ctx,
		cancel:          cancel,
		grp:             grp,
		grpCtx:          grpCtx,
		shutdownTimeout: DefaultShutdownTimeout,
		keepAlive:       true,
		log:             slog.Default(),
	}

	for _, f := range opts {
		f(ret)
	}

	var shutdownCancel context.CancelFunc
	ret.shutdownCtx, shutdownCancel = context.WithTimeout(context.Background(), ret.shutdownTimeout)
	ret.shutdownGrp, _ = errgroup.WithContext(ret.shutdownCtx)

	shutdownHandler := func() {
		<-grpCtx.Done()

		ret.lock.Lock()
		ret.state = stopping
		ret.lock.Unlock()

		for _, f := range ret.shutdownQ {
			ret.shutdownGrp.Go(f)
		}

		go func() {
			defer shutdownCancel()
			err := ret.shutdownGrp.Wait()
			if err != nil {
				// pass the base ctx here because the logger will use it to add values to the output
				ret.log.ErrorContext(ret.ctx, fmt.Sprintf("error encountered during shutdown: %v", err))
			}

			ret.lock.Lock()
			ret.state = terminated
			ret.lock.Unlock()
		}()

		<-ret.shutdownCtx.Done()
	}

	if ret.keepAlive {
		// queuing this in the pool will keep the group open until something returns an error
		ret.grp.Go(func() error {
			shutdownHandler()
			return nil
		})
	} else {
		// if we don't want keepAlive, run outside of the pool instead
		go shutdownHandler()
	}

	return ret
}

// Go schedules a task (f) to run as apart of this app.
// The task will be passed a context that, once canceled,
// indicates that the app is entering is Stopping phase
// and this task should exit. If it does not, it will
// be forcibly terminated after the stopping grace period
// has passed.
func (a *App) Go(f func(context.Context) error) error {
	// read-lock okay because grp is thread safe
	a.lock.RLock()
	defer a.lock.RUnlock()

	if err := a.ensureValidState(runMainJob, running); err != nil {
		return err
	}

	a.grp.Go(func() error { return f(a.grpCtx) })
	return nil
}

// OnShutdown queues a task to run once this app
// begins entering its shutdown phase.
func (a *App) OnShutdown(f func() error) error {
	// write-lock needed here because shutdownQ is modified
	a.lock.Lock()
	defer a.lock.Unlock()

	if err := a.ensureValidState(registerCleanupJob, running, stopping); err != nil {
		return err
	}

	switch a.state {
	case running:
		// if were running, queue for later
		a.shutdownQ = append(a.shutdownQ, f)
	case stopping:
		// if we're already stopping, run it now
		a.shutdownGrp.Go(f)
	}

	return nil
}

// Wait blocks until the shutdown phase of this app has completed
// (gracefully or forcibly), then returns the any error that occurred
// from tasks queued by calling [App.Go] during the Running phase.
func (a *App) Wait() error {
	err := a.grp.Wait()
	<-a.shutdownCtx.Done()
	return err
}

func (a *App) ensureValidState(actionName actionName, validStates ...state) error {
	if slices.Contains(validStates, a.state) {
		return nil
	}

	msg := fmt.Sprintf("%q failed. (invalid state: %v)", string(actionName), a.state)
	a.log.WarnContext(a.ctx, msg)
	return errors.New(msg)
}

// AppOptFunc is an option's function that can be used to override
// default values of an app during initialization by [New].
type AppOptFunc func(*App)

// WithShutdownTimeout sets the shutdown grace period to timeout.
func WithShutdownTimeout(timeout time.Duration) AppOptFunc {
	return func(a *App) {
		a.shutdownTimeout = timeout
	}
}

// WithLogger sets the app's logger.
func WithLogger(logger *slog.Logger) AppOptFunc {
	return func(a *App) {
		a.log = logger
	}
}

// WithKeepAlive sets the app's keep-alive value.
func WithKeepAlive(keepAlive bool) AppOptFunc {
	return func(a *App) {
		a.keepAlive = keepAlive
	}
}
