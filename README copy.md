# lanhub

A reusable Go LAN event hub built on Socket.IO. This is the server-library milestone; it has no Wails dependency. QR/PIN pairing and the Wails desktop application are intentionally left for the next milestone.

## Requirements

- Go 1.26+
- Socket.IO JavaScript client 4.x

This project uses `github.com/zishang520/socket.io/servers/socket/v3` v3.0.6. The upstream v3 line requires Go 1.26 or newer.

## Install

```bash
go get github.com/yourname/lanhub
```

Replace `github.com/yourname/lanhub` with the module path you publish.

## Quick start

```go
package main

import (
    "log"
    "github.com/yourname/lanhub"
)

func main() {
    hub := lanhub.New(lanhub.Options{
        Host: "0.0.0.0",
        Port: 8080,
    })

    hub.OnConnect(func(device *lanhub.Device) {
        log.Println("connected:", device.ID)
    })

    hub.On("message", func(ctx *lanhub.Context) {
        log.Println(ctx.Device.ID, ctx.Data)
        _ = ctx.Emit("message:received", map[string]any{"ok": true})
    })

    if err := hub.Start(); err != nil {
        log.Fatal(err)
    }
    select {}
}
```

## Options

| Option | Default | Description |
|---|---:|---|
| `Host` | `0.0.0.0` | Listen interface |
| `Port` | `8080` | TCP port |
| `Path` | `/socket.io/` | Socket.IO endpoint |
| `MaxDevices` | `20` | Maximum simultaneous devices; `-1` disables the limit |
| `PingInterval` | `25s` | Reserved Engine.IO heartbeat setting |
| `PingTimeout` | `20s` | Reserved Engine.IO heartbeat setting |
| `AllowOrigins` | `[]` | Empty permits any browser origin; otherwise exact origin allow-list |

## API

- `New(options ...Options) *Server`
- `On(event string, handler Handler) *Server`
- `OnConnect(func(*Device)) *Server`
- `OnDisconnect(func(*Device)) *Server`
- `Start() error`
- `Stop(context.Context) error`
- `Running() bool`
- `Addr() string`
- `URL() string`
- `Devices() []*Device`
- `Device(id string) (*Device, bool)`
- `Emit(event string, data ...any) error`
- `EmitTo(deviceID, event string, data ...any) error`
- `Join(deviceID, room string) error`
- `Leave(deviceID, room string) error`
- `EmitToRoom(room, event string, data ...any) error`

`Context` provides `Device`, `Data`, `Emit`, `Broadcast`, `Join` and `Leave`.

## Client protocol

Connect using the standard Socket.IO client:

```js
import { io } from "socket.io-client";

const socket = io("http://192.168.1.10:8080", {
  path: "/socket.io/",
  transports: ["websocket", "polling"]
});

socket.on("lanhub:ready", info => {
  console.log("assigned device id:", info.deviceId);
});

socket.emit("device:identify", {
  name: "My phone",
  platform: "android",
  metadata: { model: "example" }
});

socket.emit("message", { text: "hello" });
```

The server sends `lanhub:ready` after connection. The client may send `device:identify` to set its display name, platform and metadata. Each connected Socket.IO socket is assigned a device ID equal to its Socket.IO socket ID.

## Run the example

Terminal 1:

```bash
go run ./examples/basic
```

Terminal 2:

```bash
cd examples/client
pnpm install
pnpm start
```

Use the host computer's LAN IP instead of `127.0.0.1` when connecting from another device.

## Notes before production

- This first milestone does not implement pairing, PINs, QR codes, persistent device identities, TLS, authorization or a Wails UI.
- `device:identify` is descriptive metadata, not authentication. Do not trust it for authorization.
- Restrict `AllowOrigins` when serving browser clients. CORS is not an authentication mechanism.
- The current event handlers are intended to be registered before `Start()`.
- `PingInterval` and `PingTimeout` are included in the public options for the planned Engine.IO configuration; this initial adapter uses upstream defaults.
- `MaxDevices` is a connection cap, not a security boundary.
- For use beyond a trusted LAN, add pairing tokens and TLS before exposing the server.

## License

MIT
