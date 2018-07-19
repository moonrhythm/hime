package hime

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefulShutdownApp is the app in graceful shutdown mode
type GracefulShutdownApp struct {
	*gracefulShutdown

	App *App
}

type gracefulShutdown struct {
	timeout time.Duration
	wait    time.Duration
	notiFns []func()
}

// Address sets server address
func (gs *GracefulShutdownApp) Address(addr string) *GracefulShutdownApp {
	gs.App.Addr = addr
	return gs
}

// Timeout sets shutdown timeout for graceful shutdown,
// set to 0 to disable timeout
//
// default is 0
func (gs *GracefulShutdownApp) Timeout(d time.Duration) *GracefulShutdownApp {
	gs.timeout = d
	return gs
}

// Wait sets wait time before shutdown
func (gs *GracefulShutdownApp) Wait(d time.Duration) *GracefulShutdownApp {
	gs.wait = d
	return gs
}

// Notify calls fn when receive terminate signal from os
func (gs *GracefulShutdownApp) Notify(fn func()) *GracefulShutdownApp {
	if fn != nil {
		gs.notiFns = append(gs.notiFns, fn)
	}
	return gs
}

// OnShutdown calls server.RegisterOnShutdown(fn)
func (gs *GracefulShutdownApp) OnShutdown(fn func()) *GracefulShutdownApp {
	gs.App.srv.RegisterOnShutdown(fn)
	return gs
}

func (gs *GracefulShutdownApp) start(listenAndServe func() error) (err error) {
	serverCtx, cancelServer := context.WithCancel(context.Background())
	defer cancelServer()
	go func() {
		if err = listenAndServe(); err != http.ErrServerClosed {
			cancelServer()
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)

	select {
	case <-serverCtx.Done():
		return
	case <-stop:
		for _, fn := range gs.notiFns {
			go fn()
		}
		if gs.wait > 0 {
			time.Sleep(gs.wait)
		}

		if gs.timeout > 0 {
			ctx, cancel := context.WithTimeout(context.Background(), gs.timeout)
			defer cancel()
			err = gs.App.srv.Shutdown(ctx)
		} else {
			err = gs.App.srv.Shutdown(context.Background())
		}
	}
	return
}

// ListenAndServe starts web server in graceful shutdown mode
func (gs *GracefulShutdownApp) ListenAndServe() error {
	if gs.App.certFile != "" && gs.App.keyFile != "" {
		return gs.ListenAndServeTLS(gs.App.certFile, gs.App.keyFile)
	}

	return gs.start(gs.App.listenAndServe)
}

// ListenAndServeTLS starts web server in graceful shutdown and tls mode
func (gs *GracefulShutdownApp) ListenAndServeTLS(certFile, keyFile string) error {
	return gs.start(func() error { return gs.App.listenAndServeTLS(certFile, keyFile) })
}
