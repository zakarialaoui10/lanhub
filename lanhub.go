// Package lanhub provides a reusable LAN event hub built on Socket.IO.
package lanhub

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	socketio "github.com/zishang520/socket.io/servers/socket/v3"
)

var (
	ErrAlreadyRunning = errors.New("lanhub: server is already running")
	ErrNotRunning     = errors.New("lanhub: server is not running")
	ErrDeviceNotFound = errors.New("lanhub: device not found")
	ErrInvalidOptions = errors.New("lanhub: invalid server options")
)

type Handler func(*Context)

type Server struct {
	opts Options

	mu       sync.RWMutex
	handlers map[string][]Handler
	devices  map[string]*Device
	sockets  map[string]*socketio.Socket

	io       *socketio.Server
	http     *http.Server
	listener net.Listener
	running  bool
	starting bool
}

func New(options ...Options) *Server {
	opts := DefaultOptions()
	if len(options) > 0 {
		opts = options[0].withDefaults()
	}
	return &Server{
		opts:     opts,
		handlers: make(map[string][]Handler),
		devices:  make(map[string]*Device),
		sockets:  make(map[string]*socketio.Socket),
	}
}

func (s *Server) Options() Options { return s.opts }

func (s *Server) On(event string, handler Handler) *Server {
	if event == "" || handler == nil {
		return s
	}
	s.mu.Lock()
	s.handlers[event] = append(s.handlers[event], handler)
	s.mu.Unlock()
	return s
}

func (s *Server) OnConnect(handler func(*Device)) *Server {
	if handler == nil {
		return s
	}
	return s.On(eventConnect, func(ctx *Context) { handler(ctx.Device.Clone()) })
}

func (s *Server) OnDisconnect(handler func(*Device)) *Server {
	if handler == nil {
		return s
	}
	return s.On(eventDisconnect, func(ctx *Context) { handler(ctx.Device.Clone()) })
}

func (s *Server) Start() error {
	s.mu.Lock()
	if s.running || s.starting {
		s.mu.Unlock()
		return ErrAlreadyRunning
	}
	if err := s.opts.Validate(); err != nil {
		s.mu.Unlock()
		return fmt.Errorf("%w: %v", ErrInvalidOptions, err)
	}
	s.starting = true
	s.mu.Unlock()

	address := net.JoinHostPort(s.opts.Host, strconv.Itoa(s.opts.Port))
	listener, err := net.Listen("tcp", address)
	if err != nil {
		s.mu.Lock()
		s.starting = false
		s.mu.Unlock()
		return fmt.Errorf("lanhub: listen %s: %w", address, err)
	}

	io := socketio.NewServer(nil, nil)
	io.SetPath(s.opts.Path)
	io.On("connection", func(args ...any) {
		if len(args) == 0 {
			return
		}
		client, ok := args[0].(*socketio.Socket)
		if !ok || client == nil {
			return
		}
		s.handleConnection(client)
	})

	handler := s.cors(io.ServeHandler(nil))
	httpServer := &http.Server{
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	s.mu.Lock()
	s.io = io
	s.http = httpServer
	s.listener = listener
	s.running = true
	s.starting = false
	s.mu.Unlock()

	go func() {
		err := httpServer.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
		}
	}()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return ErrNotRunning
	}
	httpServer, io := s.http, s.io
	s.running = false
	s.mu.Unlock()

	if io != nil {
		closed := make(chan struct{})
		io.Close(func(error) { close(closed) })
		select {
		case <-closed:
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}

	err := httpServer.Shutdown(ctx)

	s.mu.Lock()
	s.io = nil
	s.http = nil
	s.listener = nil
	s.sockets = make(map[string]*socketio.Socket)
	for _, d := range s.devices {
		d.Connected = false
	}
	s.mu.Unlock()

	return err
}

func (s *Server) Running() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *Server) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}

func (s *Server) URL() string {
	addr := s.Addr()
	if addr == "" {
		return ""
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "http://" + addr
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return "http://" + net.JoinHostPort(host, port)
}

func (s *Server) Emit(event string, data ...any) error {
	s.mu.RLock()
	io, running := s.io, s.running
	s.mu.RUnlock()
	if !running || io == nil {
		return ErrNotRunning
	}
	return io.To().Emit(event, data...)
}

func (s *Server) EmitTo(deviceID, event string, data ...any) error {
	s.mu.RLock()
	client := s.sockets[deviceID]
	s.mu.RUnlock()
	if client == nil {
		return ErrDeviceNotFound
	}
	return client.Emit(event, data...)
}

func (s *Server) Join(deviceID, room string) error {
	s.mu.RLock()
	client := s.sockets[deviceID]
	s.mu.RUnlock()
	if client == nil {
		return ErrDeviceNotFound
	}
	client.Join(socketio.Room(room))
	return nil
}

func (s *Server) Leave(deviceID, room string) error {
	s.mu.RLock()
	client := s.sockets[deviceID]
	s.mu.RUnlock()
	if client == nil {
		return ErrDeviceNotFound
	}
	client.Leave(socketio.Room(room))
	return nil
}

func (s *Server) EmitToRoom(room, event string, data ...any) error {
	s.mu.RLock()
	io, running := s.io, s.running
	s.mu.RUnlock()
	if !running || io == nil {
		return ErrNotRunning
	}
	return io.To(socketio.Room(room)).Emit(event, data...)
}

func (s *Server) Devices() []*Device {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Device, 0, len(s.devices))
	for _, device := range s.devices {
		out = append(out, device.Clone())
	}
	return out
}

func (s *Server) Device(id string) (*Device, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.devices[id]
	if !ok {
		return nil, false
	}
	return d.Clone(), true
}

func (s *Server) handleConnection(client *socketio.Socket) {
	id := string(client.Id())
	device := &Device{
		ID:          id,
		Name:        "Unknown device",
		Platform:    "unknown",
		Connected:   true,
		ConnectedAt: time.Now().UTC(),
	}

	s.mu.Lock()
	if s.opts.MaxDevices > 0 && len(s.sockets) >= s.opts.MaxDevices {
		s.mu.Unlock()
		client.Emit("lanhub:error", map[string]any{"code": "device_limit", "message": "maximum connected devices reached"})
		client.Disconnect(true)
		return
	}
	s.sockets[id] = client
	s.devices[id] = device
	events := make([]string, 0, len(s.handlers))
	for event := range s.handlers {
		if event != eventConnect && event != eventDisconnect {
			events = append(events, event)
		}
	}
	s.mu.Unlock()

	client.Emit("lanhub:ready", map[string]any{
		"deviceId": id,
		"server":   "lanhub",
		"version":  Version,
	})

	client.On("device:identify", func(args ...any) {
		if len(args) == 0 {
			return
		}
		payload, ok := args[0].(map[string]any)
		if !ok {
			return
		}
		s.mu.Lock()
		if current := s.devices[id]; current != nil {
			if name, ok := payload["name"].(string); ok && name != "" {
				current.Name = name
			}
			if platform, ok := payload["platform"].(string); ok && platform != "" {
				current.Platform = platform
			}
			if metadata, ok := payload["metadata"].(map[string]any); ok {
				current.Metadata = metadata
			}
		}
		s.mu.Unlock()
	})

	for _, event := range events {
		s.bindEvent(client, device, event)
	}

	s.dispatch(eventConnect, &Context{Device: device.Clone(), socket: client})

	client.On("disconnect", func(...any) {
		s.mu.Lock()
		if current := s.devices[id]; current != nil {
			current.Connected = false
		}
		delete(s.sockets, id)
		current := s.devices[id]
		s.mu.Unlock()
		if current != nil {
			s.dispatch(eventDisconnect, &Context{Device: current.Clone()})
		}
	})
}

func (s *Server) bindEvent(client *socketio.Socket, device *Device, event string) {
	client.On(event, func(args ...any) {
		var data any
		if len(args) == 1 {
			data = args[0]
		} else if len(args) > 1 {
			data = args
		}
		ctx := &Context{
			Device: device.Clone(),
			Data:   data,
			socket: client,
		}
		s.dispatch(event, ctx)
	})
}

func (s *Server) dispatch(event string, ctx *Context) {
	s.mu.RLock()
	handlers := append([]Handler(nil), s.handlers[event]...)
	s.mu.RUnlock()
	for _, handler := range handlers {
		handler(ctx)
	}
}

func (s *Server) cors(next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(s.opts.AllowOrigins))
	for _, origin := range s.opts.AllowOrigins {
		allowed[origin] = struct{}{}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			if len(allowed) == 0 {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Add("Vary", "Origin")
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

const (
	eventConnect    = "$connect"
	eventDisconnect = "$disconnect"
)
