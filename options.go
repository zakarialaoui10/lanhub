package lanhub

import (
	"fmt"
	"strings"
	"time"
)

type Options struct {
	Host         string
	Port         int
	Path         string
	MaxDevices   int
	PingInterval time.Duration
	PingTimeout  time.Duration
	AllowOrigins []string
}

func DefaultOptions() Options {
	return Options{
		Host:         "0.0.0.0",
		Port:         8080,
		Path:         "/socket.io/",
		MaxDevices:   20,
		PingInterval: 25 * time.Second,
		PingTimeout:  20 * time.Second,
	}
}

func (o Options) withDefaults() Options {
	d := DefaultOptions()
	if o.Host == "" { o.Host = d.Host }
	if o.Port == 0 { o.Port = d.Port }
	if o.Path == "" { o.Path = d.Path }
	if o.MaxDevices == 0 { o.MaxDevices = d.MaxDevices }
	if o.PingInterval == 0 { o.PingInterval = d.PingInterval }
	if o.PingTimeout == 0 { o.PingTimeout = d.PingTimeout }
	return o
}

func (o Options) Validate() error {
	if o.Port < 1 || o.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if strings.TrimSpace(o.Host) == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if !strings.HasPrefix(o.Path, "/") {
		return fmt.Errorf("Socket.IO path must start with '/'")
	}
	if o.MaxDevices < -1 {
		return fmt.Errorf("MaxDevices must be -1 (unlimited) or greater")
	}
	if o.PingInterval < 0 || o.PingTimeout < 0 {
		return fmt.Errorf("ping durations cannot be negative")
	}
	return nil
}
