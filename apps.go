package hime

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
)

// Apps is the collection of App to start together
type Apps struct {
	list []*App
	gs   *GracefulShutdown
}

// AppsConfig is the hime multiple apps config
type AppsConfig struct {
	GracefulShutdown *GracefulShutdown `yaml:"gracefulShutdown" json:"gracefulShutdown"`
	HTTPSRedirect    *HTTPSRedirect    `yaml:"httpsRedirect" json:"httpsRedirect"`
}

// Merge merges multiple *App into *Apps
func Merge(apps ...*App) *Apps {
	return &Apps{list: apps}
}

// Config merges config into apps config
func (apps *Apps) Config(config AppsConfig) *Apps {
	if config.GracefulShutdown != nil {
		apps.gs = config.GracefulShutdown
	}

	if rd := config.HTTPSRedirect; rd != nil {
		go func() {
			err := StartHTTPSRedirectServer(rd.Addr)
			if err != nil {
				panicf("start https redirect server error; %v", err)
			}
		}()
	}

	return nil
}

// ParseConfig parses config data
func (apps *Apps) ParseConfig(data []byte) *Apps {
	var config AppsConfig
	err := yaml.Unmarshal(data, &config)
	if err != nil {
		panicf("can not parse config; %v", err)
	}
	return apps.Config(config)
}

// ParseConfigFile parses config from file
func (apps *Apps) ParseConfigFile(filename string) *Apps {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panicf("can not read config from file; %v", err)
	}
	return apps.ParseConfig(data)
}

func (apps *Apps) listenAndServe() error {
	eg := errgroup.Group{}

	for _, app := range apps.list {
		eg.Go(app.ListenAndServe)
	}

	return eg.Wait()
}

// ListenAndServe starts web servers
func (apps *Apps) ListenAndServe() error {
	if apps.gs != nil {
		return apps.listenAndServeGracefully()
	}

	return apps.listenAndServe()
}

// GracefulShutdown changes apps to graceful shutdown mode
func (apps *Apps) GracefulShutdown() *GracefulShutdown {
	if apps.gs == nil {
		apps.gs = &GracefulShutdown{}
	}

	return apps.gs
}

// ListenAndServe starts web servers in graceful shutdown mode
func (apps *Apps) listenAndServeGracefully() error {
	errChan := make(chan error)

	go func() {
		err := apps.listenAndServe()
		if err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)

	select {
	case err := <-errChan:
		return err
	case <-stop:
		for _, fn := range apps.gs.notiFns {
			go fn()
		}
		for _, app := range apps.list {
			for _, fn := range app.gs.notiFns {
				go fn()
			}
		}

		if apps.gs.wait > 0 {
			time.Sleep(apps.gs.wait)
		}

		eg := errgroup.Group{}
		ctx := context.Background()

		if apps.gs.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, apps.gs.timeout)
			defer cancel()
		}

		for _, app := range apps.list {
			app := app
			eg.Go(func() error { return app.srv.Shutdown(ctx) })
		}

		return eg.Wait()
	}
}
