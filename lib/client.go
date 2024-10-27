package lib

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"time"

	"github.com/syncthing/syncthing/lib/discover"
	"github.com/syncthing/syncthing/lib/events"
	"github.com/syncthing/syncthing/lib/protocol"
	relayclient "github.com/syncthing/syncthing/lib/relay/client"
	"github.com/syncthing/syncthing/lib/tlsutil"
)

func LookupDevice(ctx context.Context, deviceID protocol.DeviceID) (string, error) {
	var cert tls.Certificate
	client, err := discover.NewGlobal(LookupAddr, cert, nil, events.NoopLogger)
	if err != nil {
		return "", err
	}

	addrs, err := client.Lookup(ctx, deviceID)
	if err != nil {
		return "", err
	}

	if len(addrs) == 0 {
		return "", fmt.Errorf("no address found")
	}

	return addrs[0], nil
}

func Connect(logger *log.Logger, ctx context.Context, clientName string, deviceID protocol.DeviceID, addr string) (*tls.Conn, error) {
	cert, err := tlsutil.NewCertificateInMemory(clientName, 20*365)
	if err != nil {
		return nil, err
	}

	relayuri, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}

	logger.Println("connecting to", relayuri, ":", deviceID)
	invite, err := relayclient.GetInvitationFromRelay(ctx, relayuri, deviceID, []tls.Certificate{cert}, ConnTimeout)
	if err != nil {
		return nil, err
	}

	logger.Println("joining session", invite)
	conn, err := relayclient.JoinSession(ctx, invite)
	if err != nil {
		return nil, err
	}

	tlsConfig := TlsConfig(cert)
	sconn := tls.Client(conn, tlsConfig)

	logger.Println("performing tls handshake for", invite)
	err = PerformHandshakeAndValidation(sconn, deviceID)
	if err != nil {
		sconn.Close()
		return nil, err
	}

	return sconn, nil
}

func LookupAndConnect(logger *log.Logger, ctx context.Context, clientName string, deviceID protocol.DeviceID) (*tls.Conn, error) {
	logger.Println("looking up", deviceID)
	addr, err := LookupDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	logger.Println("found", deviceID, "at", addr)

	i := 0
	for {
		ret, err := Connect(logger, ctx, clientName, deviceID, addr)
		if err == nil || i >= 10 {
			return ret, err
		}
		i += 1
		logger.Println("error:", err)
		logger.Println("retry", i, "of 10")
	}
}

func Ping(logger *log.Logger, conn *tls.Conn) error {
	defer conn.SetDeadline(time.Time{})
	buffer := make([]byte, 1)

	buffer[0] = MsgPing
	conn.SetDeadline(time.Now().Add(ConnTimeout))
	n, err := conn.Write(buffer)
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("unexpected ping send length: %d", n)
	}
	logger.Println("sent ping")

	conn.SetDeadline(time.Now().Add(ConnTimeout))
	n, err = io.ReadFull(conn, buffer)
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("unexpected pong recv length: %d", n)
	}
	if buffer[0] != MsgPong {
		return fmt.Errorf("unexpected pong byte: %d", buffer[0])
	}
	logger.Println("recv pong")

	return nil
}

func Proxy(sconn *tls.Conn) (net.Conn, error) {
	defer sconn.SetDeadline(time.Time{})
	buffer := make([]byte, 1)

	buffer[0] = MsgProxy
	sconn.SetDeadline(time.Now().Add(ConnTimeout))
	n, err := sconn.Write(buffer)
	if err != nil {
		return nil, err
	}
	if n != 1 {
		return nil, fmt.Errorf("unexpected proxy send length: %d", n)
	}

	return sconn.NetConn(), nil
}
