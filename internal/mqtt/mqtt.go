package mqtt

import (
	"context"
	"net/url"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"

	conf "github.com/geoff-coppertop/wunderground-uploader/internal/config"
	log "github.com/sirupsen/logrus"
)

func Connect(ctx context.Context, cfg conf.Config, handler paho.MessageHandler) (*autopaho.ConnectionManager, error) {
	clientCfg := autopaho.ClientConfig{
		BrokerUrls:        []*url.URL{cfg.ServerURL},
		KeepAlive:         cfg.KeepAlive,
		ConnectRetryDelay: cfg.ConnectRetryDelay,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			log.Info("mqtt connection up")

			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: map[string]paho.SubscribeOptions{
					cfg.Topic: {},
				},
			}); err != nil {
				log.Errorf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
				return
			}

			log.Info("mqtt subscription made")
		},
		OnConnectError: func(err error) { log.Errorf("error whilst attempting connection: %s\n", err) },
		ClientConfig: paho.ClientConfig{
			Router:        paho.NewSingleHandlerRouter(handler),
			OnClientError: func(err error) { log.Errorf("server requested disconnect: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					log.Errorf("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					log.Errorf("server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	}

	// Connect to the broker
	return autopaho.NewConnection(ctx, clientCfg)
}

func Disconnect(cm *autopaho.ConnectionManager) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return cm.Disconnect(ctx)
}
