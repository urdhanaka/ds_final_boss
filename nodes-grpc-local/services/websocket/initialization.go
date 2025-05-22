package websocket

const (
	WEBSOCKET_ADDRESS       = "localhost:8080"
	WEBSOCKET_CHAN_MAX_SIZE = 30
)

type Websocket struct {
	HostnameChan map[string]chan string
}

func NewWebsocket() *Websocket {
	return &Websocket{
		HostnameChan: map[string]chan string{},
	}
}
