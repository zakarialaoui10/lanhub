package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zakarialaoui10/lanhub"
)

func main() {
	hub := lanhub.New(lanhub.Options{
		Host: "0.0.0.0",
		Port: 8080,
	})

	hub.OnConnect(func(device *lanhub.Device) {
		log.Printf("connected: %s (%s)", device.ID, device.Name)
	})

	hub.OnDisconnect(func(device *lanhub.Device) {
		log.Printf("disconnected: %s", device.ID)
	})

	hub.On("chat:message", func(ctx *lanhub.Context) {
		log.Printf("message from %s: %#v", ctx.Device.ID, ctx.Data)
		_ = ctx.Emit("chat:received", map[string]any{"ok": true})
	})

	if err := hub.Start(); err != nil {
		log.Fatal(err)
	}
	log.Printf("lanhub listening at %s", hub.Addr())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := hub.Stop(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
