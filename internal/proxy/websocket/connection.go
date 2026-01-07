package websocket

import (
	"sync"

	"github.com/gorilla/websocket"

	"github.com/xgaicc/binance-proxy/internal/logging"
)

type ConnectionProxy struct {
	client   *websocket.Conn
	server   *websocket.Conn
	logger   *logging.RequestLogger
	clientIP string
	apiType  string
	done     chan struct{}
	once     sync.Once
}

func NewConnectionProxy(
	client, server *websocket.Conn,
	logger *logging.RequestLogger,
	clientIP, apiType string,
) *ConnectionProxy {
	return &ConnectionProxy{
		client:   client,
		server:   server,
		logger:   logger,
		clientIP: clientIP,
		apiType:  apiType,
		done:     make(chan struct{}),
	}
}

func (p *ConnectionProxy) Start() {
	// Client -> Server
	go p.forward(p.client, p.server, "client->server")

	// Server -> Client
	go p.forward(p.server, p.client, "server->client")

	<-p.done
}

func (p *ConnectionProxy) forward(src, dst *websocket.Conn, direction string) {
	defer p.close()

	for {
		messageType, message, err := src.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				p.logger.Debug("WebSocket read completed",
					logging.Field("direction", direction),
					logging.Field("client_ip", p.clientIP))
			}
			return
		}

		// Log the message
		p.logger.LogWebSocketMessage(direction, p.clientIP, p.apiType, message)

		if err := dst.WriteMessage(messageType, message); err != nil {
			p.logger.Debug("WebSocket write completed",
				logging.Field("direction", direction),
				logging.Field("client_ip", p.clientIP))
			return
		}
	}
}

func (p *ConnectionProxy) close() {
	p.once.Do(func() {
		close(p.done)
	})
}
