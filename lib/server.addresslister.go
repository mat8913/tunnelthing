package lib

import (
	"github.com/syncthing/syncthing/lib/discover"
)

type serverAddressLister struct {
	Server *Server
}

func (server *Server) AddressLister() discover.AddressLister {
	return serverAddressLister{Server: server}
}

func (lister serverAddressLister) AllAddresses() []string {
	return lister.ExternalAddresses()
}

func (lister serverAddressLister) ExternalAddresses() []string {
	uri := lister.Server.relay.URI()
	if uri != nil {
		return []string{uri.String()}
	} else {
		return []string{}
	}
}
