package data

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

type wsPosPub struct {
	sync.Mutex
	address, path string
	errChan       chan error
	listeners     map[*websocket.Conn]chan<- []byte
}

// WebsocketPublisher creates a new websocket server that publishes GPS
// positions to connected clients.
func WebsocketPublisher(address, path string) Publisher {
	pub := &wsPosPub{
		address:   address,
		path:      path,
		errChan:   make(chan error),
		listeners: map[*websocket.Conn]chan<- []byte{},
	}
	pub.init()
	return pub
}

// PublishPos broadcasts a GPS position to all connected clients.
func (pub *wsPosPub) Publish(bs []byte) error {
	select {
	case err := <-pub.errChan:
		close(pub.errChan)
		return err
	default:
		pub.broadcast(bs)
	}
	return nil
}

// Initializes the HTTP server. It routes the desired path to the websocket
// server. It handles the connection upgrade. For new connections, see
// handleConn.
func (pub *wsPosPub) init() {
	go func() {
		srv := websocket.Server{Handler: websocket.Handler(pub.handleConn)}
		mux := http.NewServeMux()
		mux.Handle(pub.path, srv)

		fmt.Println("Listening on", pub.address)
		pub.errChan <- http.ListenAndServe(pub.address, mux)
	}()
}

// Broadcasts a position to all client listeners
func (pub *wsPosPub) broadcast(bs []byte) {
	pub.Lock()
	defer pub.Unlock()
	for _, lst := range pub.listeners {
		lst <- bs
	}
}

// Handles new connection to the server. It registers the new connection as a
// listener for new positions and spawns a routine for receiving client
// messages.
func (pub *wsPosPub) handleConn(conn *websocket.Conn) {
	fmt.Println("Received conn")

	lstc := pub.connect(conn)
	defer pub.disconnect(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go pub.handleRcvs(ctx, cancel, conn)

	// It has to block. Returning from the function terminates all communication
	// with the client.
	pub.listen(ctx, cancel, conn, lstc)
}

// Actively listens to broadcast channel and sends the received data to the
// client every time it receives updates.
func (pub *wsPosPub) listen(ctx context.Context, cancel func(), conn *websocket.Conn, lstc <-chan []byte) {
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case bs, ok := <-lstc:
			if !ok {
				return
			}
			if err := websocket.Message.Send(conn, string(bs)); err != nil {
				fmt.Println("Failed sending data to client:", err)
				return
			}
		}
	}
}

// Handles received data from a client. All text received is ignore. It only
// looks for errors, so it nows when to disconnect. An error may singal a close
// request from the client, not only problems.
func (pub *wsPosPub) handleRcvs(ctx context.Context, cancel func(), conn *websocket.Conn) {
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := websocket.Message.Receive(conn, nil); err != nil {
				fmt.Println("Error reading clients message:", err)
				return
			}
		}
	}
}

// Tries to close a connection with a client. It doesn't matter if it fails, it
// is just a formality, for something may have gone wrong on either side of the
// connection. It unregister the connection and closes its listener channel; no
// messages will be broadcast to it anymore.
func (pub *wsPosPub) disconnect(conn *websocket.Conn) {
	conn.Close()
	pub.Lock()
	defer pub.Unlock()
	close(pub.listeners[conn])
	delete(pub.listeners, conn)
}

// Handles a new clinet connection. Register its connection and maps it to a
// channels that listens to position broadcasts. Returns the channel listener
// channel.
func (pub *wsPosPub) connect(conn *websocket.Conn) <-chan []byte {
	lst := make(chan []byte)
	pub.Lock()
	defer pub.Unlock()
	pub.listeners[conn] = lst
	return lst
}
