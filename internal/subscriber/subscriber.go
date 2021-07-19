package subscriber

import (
	"context"
	"encoding/json"
	"sync"

	cfg "github.com/geoff-coppertop/wunderground-uploader/internal/config"
	"github.com/geoff-coppertop/wunderground-uploader/internal/mqtt"
	log "github.com/sirupsen/logrus"
)

func Start(ctx context.Context, wg *sync.WaitGroup, cfg cfg.Config) <-chan map[string]interface{} {
	out := make(chan map[string]interface{})
	var con *mqtt.Connection
	var err error = nil

	if con, err = mqtt.Connect(ctx, cfg); err != nil {
		log.Debug("connect failed")
		log.Debug(err)
		close(out)
		return out
	} else {
		log.Debug("connection started")
		wg.Add(1)
	}

	go func() {
		defer close(out)
		defer wg.Done()

		log.Debug("Waiting for the connection")
		select {
		case <-con.OnConnectionUp():
			log.Debug("<-con.OnConnectionUp")

		case err := <-con.OnError():
			log.Debug("<-con.OnError")
			log.Error(err)
			con.Disconnect()
			return

		case <-ctx.Done():
			log.Debug("<-ctx.Done")
			con.Disconnect()
			return
		}
		log.Debug("connected")

		con.Subscribe(cfg.Topic)

		for {
			select {
			case err := <-con.OnError():
				log.Debug("<-con.OnError")
				log.Error(err)
				con.Disconnect()
				return

			case <-con.OnServerDisconnect():
				log.Debug("disconnect sig")
				return

			case data := <-con.OnMessage():
				var sensorData map[string]interface{}
				if err := json.Unmarshal(data.Data, &sensorData); err != nil {
					log.Error(err)
					con.Disconnect()
					return
				}

				out <- sensorData

			case <-ctx.Done():
				log.Debug("<-ctx.Done")
				con.Disconnect()
				return
			}
		}
	}()

	return out
}
