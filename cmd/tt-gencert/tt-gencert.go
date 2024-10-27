package main

import (
	"errors"
	"os"

	"github.com/syncthing/syncthing/lib/tlsutil"
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	return !errors.Is(err, os.ErrNotExist)
}

func main() {
	certFile := "cert.pem"
	keyFile := "key.pem"

	if fileExists(certFile) {
		println(certFile, "already exists")
		return
	}

	if fileExists(keyFile) {
		println(keyFile, "already exists")
		return
	}

	_, err := tlsutil.NewCertificate(certFile, keyFile, "strelaysrv", 20*365)
	if err != nil {
		panic(err)
	}
}
