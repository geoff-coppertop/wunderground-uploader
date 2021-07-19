package mqtt

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	cfg "github.com/geoff-coppertop/wunderground-uploader/internal/config"
	log "github.com/sirupsen/logrus"
)

type Data struct {
	Topic string
	Data  []byte
}

type Connection struct {
	connectionManager *autopaho.ConnectionManager
	ctx               context.Context

	mu             sync.Mutex // protects following fields
	onConnectionUp atomic.Value
	onMessage      atomic.Value // of chan Data, created lazily, closed by ...
	onError        atomic.Value
	onDisconnect   atomic.Value
}

// closedchan is a reusable closed channel.
var closedchan = make(chan struct{})

func init() {
	close(closedchan)
}

func Connect(ctx context.Context, cfg cfg.Config) (*Connection, error) {
	var err error = nil
	conn := &Connection{ctx: ctx}

	clientCfg := autopaho.ClientConfig{
		BrokerUrls:        []*url.URL{cfg.ServerURL},
		KeepAlive:         cfg.KeepAlive,
		ConnectRetryDelay: cfg.ConnectRetryDelay,

		OnConnectionUp: conn.connectionUpHandler,
		OnConnectError: conn.errorHandler,

		ClientConfig: paho.ClientConfig{
			Router:             paho.NewSingleHandlerRouter(conn.messageHandler),
			OnClientError:      conn.errorHandler,
			OnServerDisconnect: conn.serverDisconnectHandler,
		},
	}

	// Connect to the broker
	if conn.connectionManager, err = autopaho.NewConnection(ctx, clientCfg); err != nil {
		return nil, err
	}

	return conn, err
}

func (conn *Connection) OnMessage() <-chan Data {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	om := conn.onMessage.Load()

	if om == nil {
		om = make(chan Data)
		conn.onMessage.Store(om)
	}

	return om.(chan Data)
}

func (conn *Connection) OnConnectionUp() <-chan struct{} {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	ocu := conn.onConnectionUp.Load()

	if ocu == nil {
		ocu = make(chan struct{})
		conn.onConnectionUp.Store(ocu)
	}

	return ocu.(chan struct{})
}

func (conn *Connection) OnError() <-chan error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	oe := conn.onError.Load()

	if oe == nil {
		oe = make(chan error)
		conn.onError.Store(oe)
	}

	return oe.(chan error)
}

func (conn *Connection) OnServerDisconnect() <-chan struct{} {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	od := conn.onDisconnect.Load()

	if od == nil {
		od = make(chan struct{})
		conn.onDisconnect.Store(od)
	}

	return od.(chan struct{})
}

func (conn *Connection) Subscribe(topic string) {
	// Subscribe may block so we run it in a goRoutine
	go func() {
		ctx, cancel := context.WithTimeout(conn.ctx, 100*time.Millisecond)
		defer cancel()

		// AwaitConnection will return immediately if connection is up; adding this call stops subscription whilst
		// connection is unavailable.
		if err := conn.connectionManager.AwaitConnection(ctx); err != nil {
			// Should only happen when context is cancelled
			log.Errorf("Subscribe timed out waiting for the connection: %v", err)
			conn.errorHandler(err)
			return
		}

		ctx, cancel = context.WithTimeout(conn.ctx, 100*time.Millisecond)
		defer cancel()

		if _, err := conn.connectionManager.Subscribe(ctx, &paho.Subscribe{
			Subscriptions: map[string]paho.SubscribeOptions{
				topic: {},
			},
		}); err != nil {
			log.Errorf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
			conn.errorHandler(err)
			return
		}

		log.Debug("mqtt subscription made")
	}()
}

func (conn *Connection) Publish(data Data) {
	// Publish will block so we run it in a goRoutine
	go func() {
		ctx, cancel := context.WithTimeout(conn.ctx, 100*time.Millisecond)
		defer cancel()

		// AwaitConnection will return immediately if connection is up; adding this call stops publication whilst
		// connection is unavailable.
		if err := conn.connectionManager.AwaitConnection(ctx); err != nil {
			// Should only happen when context is cancelled
			log.Errorf("Publish timed out waiting for the connection: %v", err)
			conn.errorHandler(err)
		}

		ctx, cancel = context.WithTimeout(conn.ctx, 100*time.Millisecond)
		defer cancel()

		if pr, err := conn.connectionManager.Publish(ctx, &paho.Publish{
			Topic:   data.Topic,
			Payload: data.Data,
		}); err != nil {
			log.Errorf("error publishing: %v", err)
			conn.errorHandler(err)
			return
		} else if pr != nil && pr.ReasonCode != 0 && pr.ReasonCode != 16 {
			// 16 = Server received message but there are no subscribers
			log.Debugf("reason code %d received", pr.ReasonCode)
		}

		log.Debugf("sent: %v", data)
	}()
}

func (conn *Connection) Disconnect() error {
	ctx, cancel := context.WithTimeout(conn.ctx, time.Second)
	defer cancel()

	return conn.connectionManager.Disconnect(ctx)
}

func SanitizeTopic(topic string) string {
	topic = strings.ReplaceAll(topic, " ", "_")
	topic = strings.ReplaceAll(topic, ".", "_")
	topic = strings.ReplaceAll(topic, "&", "_")

	return topic
}

func JoinTopic(topics ...string) string {
	return SanitizeTopic(strings.Join(topics, "/"))
}

func (conn *Connection) messageHandler(m *paho.Publish) {
	conn.mu.Lock()

	om := conn.onMessage.Load()

	if om == nil {
		om = make(chan Data)
		conn.onMessage.Store(om)
	}

	conn.mu.Unlock()

	om.(chan Data) <- Data{Topic: m.Topic, Data: m.Payload}
}

func (conn *Connection) connectionUpHandler(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	ocu := conn.onConnectionUp.Load()

	if ocu == nil {
		conn.onConnectionUp.Store(closedchan)
	} else {
		close(ocu.(chan struct{}))
	}
}

func (conn *Connection) errorHandler(err error) {
	conn.mu.Lock()

	oe := conn.onError.Load()

	if oe == nil {
		oe = make(chan error)
		conn.onError.Store(oe)
	}

	conn.mu.Unlock()

	oe.(chan error) <- err
}

func (conn *Connection) serverDisconnectHandler(d *paho.Disconnect) {
	if d.Properties != nil {
		log.Errorf("server requested disconnect: %s\n", d.Properties.ReasonString)
	} else {
		log.Errorf("server requested disconnect; reason code: %d\n", d.ReasonCode)
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()

	od := conn.onDisconnect.Load().(chan struct{})
	if od == nil {
		conn.onMessage.Store(closedchan)
	} else {
		close(od)
	}

	oe := conn.onError.Load().(chan error)
	if oe == nil {
		conn.onError.Store(closedchan)
	} else {
		close(oe)
	}

	om := conn.onMessage.Load().(chan Data)
	if om == nil {
		conn.onMessage.Store(closedchan)
	} else {
		close(om)
	}
}
