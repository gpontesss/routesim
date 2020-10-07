package pospub

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gpontesss/routesim/pkg/gps"
	geojson "github.com/paulmach/go.geojson"
	"golang.org/x/net/websocket"
)

type wsPosPub struct {
	sync.Mutex
	address, path string
	errChan       chan error
	listeners     map[*websocket.Conn]chan<- gps.Position
}

// WebsocketPosPublisher docs here
func WebsocketPosPublisher(address, path string) PosPublisher {
	pub := &wsPosPub{
		address:   address,
		path:      path,
		errChan:   make(chan error),
		listeners: map[*websocket.Conn]chan<- gps.Position{},
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

func (p *wsPosPub) handleConn(conn *websocket.Conn) {
	rcv := p.createListener(conn)
	fmt.Println("Received conn")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		var s string
		if err := websocket.Message.Receive(conn, &s); err != nil {
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
			pos := <-rcv
			// Should be temporary
			geoJSON := geojson.Feature{
				ID: pos.GPS.ID(),
				Geometry: &geojson.Geometry{
					Type: geojson.GeometryPoint,
					Point: []float64{
						pos.Lat.Degrees(),
						pos.Lng.Degrees(),
					},
				},
			}

			if err := websocket.JSON.Send(conn, geoJSON); err != nil {
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

func (p *wsPosPub) createListener(conn *websocket.Conn) <-chan gps.Position {
	lst := make(chan gps.Position)

	p.Lock()
	defer p.Unlock()
	p.listeners[conn] = lst
	return lst
}

func (p *wsPosPub) PublishPos(pos gps.Position) error {
	fmt.Println(pos.GPS.ID(), "Received pos", pos.LatLng, "channels", len(p.listeners))

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
