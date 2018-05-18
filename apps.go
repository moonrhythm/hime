package hime

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Apps is the collection of App to start together
type Apps struct {
	*gracefulShutdown

	list []*App
}

// Merge merges multiple *App into *Apps
func Merge(apps ...*App) *Apps {
	return &Apps{list: apps}
}

// ListenAndServe starts web servers
func (apps *Apps) ListenAndServe() error {
	wg := &sync.WaitGroup{}
	doneChan := make(chan struct{})
	errChan := make(chan error)
	for _, app := range apps.list {
		app := app
		wg.Add(1)
		go func() {
			err := app.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				errChan <- err
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		doneChan <- struct{}{}
	}()

	select {
	case err := <-errChan:
		return err
	case <-doneChan:
		return nil
	}
}

// GracefulShutdownApps is the apps in graceful shutdown mode
type GracefulShutdownApps struct {
	*gracefulShutdown

	Apps *Apps
}

// GracefulShutdown changes apps to graceful shutdown mode
func (apps *Apps) GracefulShutdown() *GracefulShutdownApps {
	if apps.gracefulShutdown == nil {
		apps.gracefulShutdown = &gracefulShutdown{}
	}
	return &GracefulShutdownApps{
		Apps:             apps,
		gracefulShutdown: apps.gracefulShutdown,
	}
}

// Timeout sets shutdown timeout for graceful shutdown,
// set to 0 to disable timeout
//
// default is 0
func (gs *GracefulShutdownApps) Timeout(d time.Duration) *GracefulShutdownApps {
	gs.timeout = d
	return gs
}

// Wait sets wait time before shutdown
func (gs *GracefulShutdownApps) Wait(d time.Duration) *GracefulShutdownApps {
	gs.wait = d
	return gs
}

// Notify calls fn when receive terminate signal from os
func (gs *GracefulShutdownApps) Notify(fn func()) *GracefulShutdownApps {
	if fn != nil {
		gs.notiFns = append(gs.notiFns, fn)
	}
	return gs
}

// ListenAndServe starts web servers in graceful shutdown mode
func (gs *GracefulShutdownApps) ListenAndServe() error {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error)
	for _, app := range gs.Apps.list {
		app := app
		wg.Add(1)
		go func() {
			err := app.listenAndServe()
			if err != nil && err != http.ErrServerClosed {
				errChan <- err
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		cancel()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-stop:
		for _, fn := range gs.notiFns {
			go fn()
		}
		for _, app := range gs.Apps.list {
			if app.gracefulShutdown != nil {
				for _, fn := range app.gracefulShutdown.notiFns {
					go fn()
				}
			}
		}
		if gs.wait > 0 {
			time.Sleep(gs.wait)
		}

		wg := &sync.WaitGroup{}
		var (
			shutdownCtx context.Context
			cancel      context.CancelFunc
		)
		if gs.timeout > 0 {
			shutdownCtx, cancel = context.WithTimeout(context.Background(), gs.timeout)
		} else {
			shutdownCtx, cancel = context.WithCancel(context.Background())
		}
		defer cancel()

		errChan := make(chan error)
		doneChan := make(chan struct{})

		for _, app := range gs.Apps.list {
			app := app
			wg.Add(1)
			go func() {
				ctx := shutdownCtx
				if app.gracefulShutdown != nil {
					if app.gracefulShutdown.wait > 0 {
						time.Sleep(app.gracefulShutdown.wait)
					}
					if app.gracefulShutdown.timeout > 0 {
						var cancel context.CancelFunc
						ctx, cancel = context.WithTimeout(shutdownCtx, app.gracefulShutdown.timeout)
						defer cancel()
					}
				}
				err := app.srv.Shutdown(ctx)
				if err != nil {
					errChan <- err
				}
				wg.Done()
			}()
		}

		go func() {
			wg.Wait()
			doneChan <- struct{}{}
		}()

		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		case <-doneChan:
			return nil
		}
	}
}
