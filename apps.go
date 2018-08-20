package hime

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
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
	eg := errgroup.Group{}

	for _, app := range apps.list {
		eg.Go(app.ListenAndServe)
	}

	return eg.Wait()
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
	eg := errgroup.Group{}

	for _, app := range gs.Apps.list {
		eg.Go(app.ListenAndServe)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)

	errChan := make(chan error)
	go func() {
		err := eg.Wait()
		if err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
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

		eg := errgroup.Group{}
		ctx := context.Background()

		if gs.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, gs.timeout)
			defer cancel()
		}

		for _, app := range gs.Apps.list {
			app := app
			eg.Go(func() error { return app.srv.Shutdown(ctx) })
		}

		return eg.Wait()
	}
}
