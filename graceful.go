package hime

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefulShutdown is the app in graceful shutdown mode
type GracefulShutdown struct {
	App       *App
	timeout   time.Duration
	wait      time.Duration
	notiFns   []func()
	beforeFns []func()
}

// Timeout sets shutdown timeout for graceful shutdown,
// set to 0 to disable timeout
//
// default is 0
func (gs *GracefulShutdown) Timeout(d time.Duration) *GracefulShutdown {
	gs.timeout = d
	return gs
}

// Wait sets wait time before shutdown
func (gs *GracefulShutdown) Wait(d time.Duration) *GracefulShutdown {
	gs.wait = d
	return gs
}

// Notify calls fn when receive terminate signal from os
func (gs *GracefulShutdown) Notify(fn func()) *GracefulShutdown {
	if fn != nil {
		gs.notiFns = append(gs.notiFns, fn)
	}
	return gs
}

// Before runs fn before start waiting to SIGTERM
func (gs *GracefulShutdown) Before(fn func()) *GracefulShutdown {
	if fn != nil {
		gs.beforeFns = append(gs.beforeFns, fn)
	}
	return gs
}

func (gs *GracefulShutdown) start(listenAndServe func() error) (err error) {
	serverCtx, cancelServer := context.WithCancel(context.Background())
	defer cancelServer()
	go func() {
		if err = listenAndServe(); err != http.ErrServerClosed {
			cancelServer()
		}
	}()

	for _, fn := range gs.beforeFns {
		fn()
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)

	select {
	case <-serverCtx.Done():
		return
	case <-stop:
		for _, fn := range gs.notiFns {
			fn()
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
func (gs *GracefulShutdown) ListenAndServe(addr string) error {
	return gs.start(func() error { return gs.App.ListenAndServe(addr) })
}

// ListenAndServeTLS starts web server in graceful shutdown and tls mode
func (gs *GracefulShutdown) ListenAndServeTLS(addr, certFile, keyFile string) error {
	return gs.start(func() error { return gs.App.ListenAndServeTLS(addr, certFile, keyFile) })
}
