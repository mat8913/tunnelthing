package lib

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"sync"

	"github.com/syncthing/syncthing/lib/discover"
	"github.com/syncthing/syncthing/lib/events"
	"github.com/syncthing/syncthing/lib/protocol"
	relayclient "github.com/syncthing/syncthing/lib/relay/client"
)

type ServerConfig struct {
	Cert          tls.Certificate
	ServerNetwork string
	ServerAddress string
}

type Server struct {
	logger *log.Logger
	ctx    context.Context
	config *ServerConfig
	relay  relayclient.RelayClient
}

func NewServer(logger *log.Logger, ctx context.Context, config *ServerConfig) (*Server, error) {
	id := protocol.NewDeviceID(config.Cert.Certificate[0])
	logger.Println("Server ID:", id)

	uri, err := url.Parse(DynamicRelayAddr)
	if err != nil {
		return nil, err
	}

	relay, err := relayclient.NewClient(uri, []tls.Certificate{config.Cert}, ConnTimeout)
	if err != nil {
		return nil, err
	}

	server := Server{
		logger: logger,
		ctx:    ctx,
		config: config,
		relay:  relay,
	}
	return &server, nil
}

func (server *Server) ServeServer() error {
	go server.relay.Serve(server.ctx)

	server.logger.Println("receiving invitations")

	for {
		select {
		case <-server.ctx.Done():
			return server.ctx.Err()
		case invite := <-server.relay.Invitations():
			debugName := fmt.Sprint(invite)
			connLogger := log.New(server.logger.Writer(), debugName+":", log.LstdFlags)
			clientID, err := protocol.DeviceIDFromBytes(invite.From)
			if err != nil {
				connLogger.Println("error getting device id:", err)
				continue
			}
			clientConn, err := relayclient.JoinSession(server.ctx, invite)
			if err != nil {
				connLogger.Println("error joining session:", err)
				continue
			}
			go func() {
				connLogger.Println("got connection")
				defer func() {
					connLogger.Println("close connection")
				}()
				err := serveConnection(connLogger, server.config, clientConn, clientID)
				if err != nil {
					connLogger.Println("error:", err)
				}
			}()
		}
	}
}

func (server *Server) ServeDiscover() error {
	client, err := discover.NewGlobal(AnnounceAddr, server.config.Cert, server.AddressLister(), events.NoopLogger)
	if err != nil {
		return err
	}

	return client.Serve(server.ctx)
}

func serveConnection(logger *log.Logger, config *ServerConfig, clientConn net.Conn, clientID protocol.DeviceID) error {
	logger.Println("serving connection")

	defer func() {
		logger.Println("closing connection")
		clientConn.Close()
	}()

	tlsConfig := TlsConfig(config.Cert)
	sconn := tls.Server(clientConn, tlsConfig)

	err := PerformHandshakeAndValidation(sconn, clientID)
	if err != nil {
		return err
	}

	buffer := make([]byte, 1)
	for {
		n, err := io.ReadFull(sconn, buffer)
		if err != nil {
			return err
		}
		if n != 1 {
			return fmt.Errorf("unexpected read length: %d")
		}

		logger.Println("got msg:", buffer[0])
		switch buffer[0] {
		case MsgPing:
			err = handlePing(logger, sconn)
			if err != nil {
				return err
			}
		case MsgProxy:
			return handleProxy(logger, config, clientConn)
		default:
			return fmt.Errorf("unknown msg: %d", buffer[0])
		}
	}
}

func handlePing(logger *log.Logger, conn *tls.Conn) error {
	buffer := make([]byte, 1)

	buffer[0] = MsgPong
	n, err := conn.Write(buffer)
	if err != nil {
		return nil
	}
	if n != 1 {
		return fmt.Errorf("unexpected pong send length: %d", n)
	}
	logger.Println("sent pong")

	return nil
}

func handleProxy(logger *log.Logger, config *ServerConfig, clientConn net.Conn) error {
	serverConn, err := net.Dial(config.ServerNetwork, config.ServerAddress)
	if err != nil {
		return err
	}

	defer clientConn.Close()
	defer serverConn.Close()

	tcpClientConn := clientConn.(*net.TCPConn)
	tcpServerConn := serverConn.(*net.TCPConn)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		defer tcpClientConn.CloseWrite()
		io.Copy(tcpClientConn, tcpServerConn)
		logger.Println("server disconnected")
	}()

	go func() {
		defer wg.Done()
		defer tcpServerConn.CloseWrite()
		io.Copy(tcpServerConn, tcpClientConn)
		logger.Println("client disconnected")
	}()

	wg.Wait()
	return nil
}
