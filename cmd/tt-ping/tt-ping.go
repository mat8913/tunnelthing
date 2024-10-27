package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/mat8913/tunnelthing/lib"
	"github.com/syncthing/syncthing/lib/protocol"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Usage: ", os.Args[0], " <DEVICE ID>")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connect := os.Args[1]

	id, err := protocol.DeviceIDFromString(connect)
	if err != nil {
		panic(err)
	}

	logger := log.New(os.Stderr, "", log.LstdFlags)

	sconn, err := lib.LookupAndConnect(logger, ctx, "tt-ping", id)
	if err != nil {
		panic(err)
	}
	defer sconn.Close()

	for {
		err = lib.Ping(logger, sconn)
		if err != nil {
			panic(err)
		}
		time.Sleep(1500 * time.Millisecond)
	}
}
