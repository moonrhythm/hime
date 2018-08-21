package hime

import (
	"encoding/json"
	"time"
)

// GracefulShutdown is the graceful shutdown configure
type GracefulShutdown struct {
	timeout time.Duration
	wait    time.Duration
	notiFns []func()
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

type gracefulConfig struct {
	Timeout string `yaml:"timeout" json:"timeout"`
	Wait    string `yaml:"wait" json:"wait"`
}

func (x *gracefulConfig) store(gs *GracefulShutdown) {
	parseDuration(x.Timeout, &gs.timeout)
	parseDuration(x.Wait, &gs.wait)
}

// UnmarshalYAML implements yaml.Unmarshaler
func (gs *GracefulShutdown) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var x gracefulConfig
	err := unmarshal(&x)
	if err != nil {
		return err
	}

	x.store(gs)

	return nil
}

// UnmarshalJSON implements json.Unmarshaler
func (gs *GracefulShutdown) UnmarshalJSON(b []byte) error {
	var x gracefulConfig
	err := json.Unmarshal(b, &x)
	if err != nil {
		return err
	}

	x.store(gs)

	return nil
}
