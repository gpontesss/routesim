# RouteSim

RouteSim simulates GPS movement and publishes the data to a resource. Current
resources are:

+ Websocket
+ AWS Kinesis (WIP)
+ Shapefile

## Getting started

To install it:

```sh
go install ./cmd/routesim
```

> For now cloning the repository is needed.
> Make sure to have $GOPATH/bin in your PATH.

There are some samples available. For example, to run a websocket publisher
sample:

```sh
routesim --config samples/websocket/websocket.json
```

Then, open [samples/websocket/index.html](samples/websocket/index.html) in your
browser. You should see moving markers representing simulated GPSs.

## Configuration

Got to describe it.
