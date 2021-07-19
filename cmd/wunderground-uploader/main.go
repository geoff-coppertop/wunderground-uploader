package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	cfg "github.com/geoff-coppertop/wunderground-uploader/internal/config"
	sub "github.com/geoff-coppertop/wunderground-uploader/internal/subscriber"
	tmg "github.com/geoff-coppertop/wunderground-uploader/internal/transmogrifier"
	wund "github.com/geoff-coppertop/wunderground-uploader/internal/wunderground"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	cfg, err := cfg.GetConfig()
	if err != nil {
		log.Panic(err)
	}

	log.SetLevel(cfg.Debug)
	log.Info("Starting")

	log.Debug(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	dataCh := sub.Start(ctx, &wg, cfg)

	wxCh := tmg.Start(ctx, &wg, dataCh)

	pubCh := wund.Start(ctx, &wg, cfg, wxCh)

	WaitProcess(&wg, pubCh, cancel)
}

func WaitProcess(wg *sync.WaitGroup, ch <-chan struct{}, cancel context.CancelFunc) {
	log.Info("Waiting")

	select {
	case <-OSExit():
		log.Info("signal caught - exiting")

	case <-ch:
		log.Errorf("uh-oh")
	}

	cancel()

	log.Info("cancelled")

	wg.Wait()

	log.Info("goodbye")
}

func OSExit() <-chan os.Signal {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)

	return sig
}
