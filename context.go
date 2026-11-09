package lanhub

import socketio "github.com/zishang520/socket.io/servers/socket/v3"

// Context represents one incoming event from a connected device.
type Context struct {
	Device *Device
	Data   any
	socket *socketio.Socket
}

// Emit sends an event to the device that triggered this handler.
func (c *Context) Emit(event string, data ...any) error {
	if c == nil || c.socket == nil {
		return ErrDeviceNotFound
	}
	return c.socket.Emit(event, data...)
}

// Broadcast sends an event to every connected device.
func (c *Context) Broadcast(event string, data ...any) error {
	if c == nil || c.socket == nil {
		return ErrNotRunning
	}
	return c.socket.Broadcast().Emit(event, data...)
}

// Join adds the sender to a Socket.IO room.
func (c *Context) Join(room string) {
	c.socket.Join(socketio.Room(room))
}

func (c *Context) Leave(room string) {
	c.socket.Leave(socketio.Room(room))
}
