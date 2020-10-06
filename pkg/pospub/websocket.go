package pospub

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	geojson "github.com/paulmach/go.geojson"
	"golang.org/x/net/websocket"
)

type wsPosPub struct {
	sync.Mutex
	address, path string
	errChan       chan error
	listeners     map[*websocket.Conn]chan<- geojson.Feature
}

// WebsocketPosPublisher docs here
func WebsocketPosPublisher(address, path string) PosPublisher {
	pub := &wsPosPub{
		address:   address,
		path:      path,
		errChan:   make(chan error),
		listeners: map[*websocket.Conn]chan<- geojson.Feature{},
	}
	pub.init()
	return pub
}

func (p *wsPosPub) init() {
	go func() {
		srv := websocket.Server{Handler: websocket.Handler(p.handleConn)}
		mux := http.NewServeMux()
		mux.Handle(p.path, srv)

		fmt.Println("Listening on", p.address)
		p.errChan <- http.ListenAndServe(p.address, mux)
	}()
}

var codec = websocket.Codec{
	Unmarshal: func(data []byte, payloadType byte, v interface{}) (err error) {
		fmt.Println(data, payloadType, v)
		if payloadType == websocket.CloseFrame {
			return errors.New("Connection closed by client")
		}
		return nil
	},
}

func (p *wsPosPub) handleConn(conn *websocket.Conn) {
	rcv := p.createListener(conn)
	fmt.Println("Received conn")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		var s string
		if err := codec.Receive(conn, &s); err != nil {
			fmt.Printf("Client disconnected: %+v\n", err)
			p.unregisterListener(conn)
			cancel()
		} else {
			fmt.Println("Received", s)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := websocket.JSON.Send(conn, <-rcv); err != nil {
				p.unregisterListener(conn)
				p.errChan <- err
				return
			}
		}
	}
}

func (p *wsPosPub) unregisterListener(conn *websocket.Conn) {
	p.Lock()
	defer p.Unlock()
	close(p.listeners[conn])
	delete(p.listeners, conn)
}

func (p *wsPosPub) createListener(conn *websocket.Conn) <-chan geojson.Feature {
	lst := make(chan geojson.Feature)

	p.Lock()
	defer p.Unlock()
	p.listeners[conn] = lst
	return lst
}

func (p *wsPosPub) PublishPos(pos geojson.Feature) error {
	fmt.Println(pos.ID, "Received pos", pos.Geometry.Point, "channels", len(p.listeners))

	select {
	case err := <-p.errChan:
		close(p.errChan)
		return err
	default:
		for _, lst := range p.listeners {
			lst <- pos
		}
	}
	return nil
}
