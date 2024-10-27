package lib

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/syncthing/syncthing/lib/protocol"
	"github.com/syncthing/syncthing/lib/tlsutil"
)

const (
	ConnTimeout = 10 * time.Second

	MsgPing  = 0x01
	MsgPong  = 0x02
	MsgProxy = 0x03

	AnnounceAddr     = "https://discovery-v6.syncthing.net/v2/?nolookup&id=LYXKCHX-VI3NYZR-ALCJBHF-WMZYSPK-QG6QJA3-MPFYMSO-U56GTUK-NA2MIAW"
	LookupAddr       = "https://discovery.syncthing.net/v2/?noannounce&id=LYXKCHX-VI3NYZR-ALCJBHF-WMZYSPK-QG6QJA3-MPFYMSO-U56GTUK-NA2MIAW"
	DynamicRelayAddr = "dynamic+https://relays.syncthing.net/endpoint"
)

func TlsConfig(cert tls.Certificate) *tls.Config {
	ret := tlsutil.SecureDefaultTLS13()
	ret.Certificates = []tls.Certificate{cert}
	ret.ClientAuth = tls.RequestClientCert
	ret.SessionTicketsDisabled = true
	ret.InsecureSkipVerify = true
	return ret
}

func PerformHandshakeAndValidation(conn *tls.Conn, deviceID protocol.DeviceID) error {
	err := tlsTimedHandshake(conn)
	if err != nil {
		return err
	}

	cs := conn.ConnectionState()
	certs := cs.PeerCertificates
	if cl := len(certs); cl != 1 {
		return fmt.Errorf("unexpected certificate count: %d", cl)
	}

	actualID := protocol.NewDeviceID(certs[0].Raw)
	if actualID != deviceID {
		return fmt.Errorf("peer id does not match. Expected %v got %v", deviceID, actualID)
	}

	return nil
}

func tlsTimedHandshake(tc *tls.Conn) error {
	tc.SetDeadline(time.Now().Add(ConnTimeout))
	defer tc.SetDeadline(time.Time{})
	return tc.Handshake()
}
