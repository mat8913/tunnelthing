package main

import (
	"context"
	"log"
	"net"
	"os"
	"reflect"
	"syscall"

	"github.com/mat8913/tunnelthing/lib"
	"github.com/syncthing/syncthing/lib/protocol"
)

// Source: https://github.com/higebu/netfd/blob/ed17b5f1ac32df732afbeeab4acd7a911e9eeacb/netfd.go
func get_conn_fd(c net.Conn) int {
	v := reflect.Indirect(reflect.ValueOf(c))
	conn := v.FieldByName("conn")
	netFD := reflect.Indirect(conn.FieldByName("fd"))
	pfd := netFD.FieldByName("pfd")
	fd := int(pfd.FieldByName("Sysfd").Int())
	return fd
}

func main() {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	if len(os.Args) != 2 {
		logger.Fatal("Usage: ", os.Args[0], " <DEVICE ID>")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deviceID := os.Args[1]

	id, err := protocol.DeviceIDFromString(deviceID)
	if err != nil {
		panic(err)
	}

	sconn, err := lib.LookupAndConnect(logger, ctx, "tt-connect", id)
	if err != nil {
		panic(err)
	}

	conn, err := lib.Proxy(sconn)
	if err != nil {
		panic(err)
	}

	fd := get_conn_fd(conn)

	rights := syscall.UnixRights(fd)
	err = syscall.Sendmsg(1, nil, rights, nil, 0)
	if err != nil {
		panic(err)
	}
}
