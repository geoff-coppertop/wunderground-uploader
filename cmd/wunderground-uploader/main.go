package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	conf "github.com/geoff-coppertop/wunderground-uploader/internal/config"
	"github.com/geoff-coppertop/wunderground-uploader/internal/mqtt"
	log "github.com/sirupsen/logrus"

	"github.com/eclipse/paho.golang/paho"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	cfg, err := conf.GetConfig()
	if err != nil {
		log.Panic(err)
	}

	log.SetLevel(cfg.Debug)
	log.Info("Starting")

	log.Debug(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cm, err := mqtt.Connect(
		ctx,
		cfg,
		func(m *paho.Publish) {
			log.Info("RX: ", m)
		})
	if err != nil {
		log.Panic(err)
	}

	// Messages will be handled through the callback so we really just need to wait until a shutdown
	// is requested
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	log.Info("Waiting")

	<-sig

	log.Info("signal caught - exiting")

	// We could cancel the context at this point but will call Disconnect instead (this waits for autopaho to shutdown)
	mqtt.Disconnect(cm)

	log.Info("shutdown complete")
}
