package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"sync"

	"github.com/mat8913/tunnelthing/lib"
	"github.com/syncthing/syncthing/lib/logger"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: ", os.Args[0], " <NETWORK> <ADDR>")
		return
	}

	logger.DefaultLogger.SetDebug("relay", true)
	logger.DefaultLogger.SetDebug("discover", true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	certFile, keyFile := "cert.pem", "key.pem"
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		panic(err)
	}

	config := &lib.ServerConfig{
		Cert:          cert,
		ServerNetwork: os.Args[1],
		ServerAddress: os.Args[2],
	}

	server, err := lib.NewServer(log.New(os.Stdout, "", log.LstdFlags), ctx, config)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = server.ServeServer()
		if err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		err = server.ServeDiscover()
		if err != nil {
			panic(err)
		}
	}()

	wg.Wait()
}
