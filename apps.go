package hime

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Apps is the collection of App to start together
type Apps struct {
	list []*App
}

// Merge merges multiple *App into *Apps
func Merge(apps ...*App) *Apps {
	return &Apps{list: apps}
}

func (apps *Apps) listenAndServe() error {
	eg, ctx := errgroup.WithContext(context.Background())

	for _, app := range apps.list {
		eg.Go(app.ListenAndServe)
	}

	<-ctx.Done()
	go apps.Shutdown()

	return eg.Wait()
}

// ListenAndServe starts web servers
func (apps *Apps) ListenAndServe() error {
	return apps.listenAndServe()
}

// Shutdown shutdowns all apps
func (apps *Apps) Shutdown() error {
	eg := errgroup.Group{}

	for _, app := range apps.list {
		app := app
		eg.Go(func() error { return app.srv.Shutdown() })
	}

	return eg.Wait()
}
