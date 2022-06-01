package main

import (
	"context"
	"tcptracker/cmd/servid"
)

func main() {
	ctx := context.Background()
	app := servid.NewApp()
	app.TrackHostConnections(ctx)
	app.ServerStart()
}
