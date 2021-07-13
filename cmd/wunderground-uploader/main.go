package main

import (
	"os"
	"os/signal"
	"syscall"

	conf "github.com/geoff-coppertop/wunderground-uploader/internal/config"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := conf.GetConfig()
	if err != nil {
		log.Panic(err)
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetLevel(cfg.Debug)
	log.Info("Starting")

	// Messages will be handled through the callback so we really just need to wait until a shutdown
	// is requested
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	log.Info("Waiting")

	<-sig

	log.Info("signal caught - exiting")

	log.Info("shutdown complete")
}
