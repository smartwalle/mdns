package mdns

import (
	"context"
	"golang.org/x/net/dns/dnsmessage"
	"net"
)

type Client interface {
	EnableIPv4()

	EnableIPv6()

	OnResource(f func(net.Addr, Resource))

	OnWarning(f func(net.Addr, error))

	OnError(f func(error))

	Start(ctx context.Context) error

	Multicast(message dnsmessage.Message) error

	Close()
}

type mClient struct {
	*mDNS
}

// NewClient creates a new object implementing the Client interface. Do not forget
// to call EnableIPv4() or EnableIPv6() to enable listening on interfaces of
// the corresponding type, or nothing will work.
func NewClient() Client {
	var nClient = &mClient{}
	nClient.mDNS = &mDNS{}
	nClient.mDNS.conn4 = nil
	nClient.mDNS.conn6 = nil
	return nClient
}

func (m *mClient) EnableIPv4() {
	var lAddr = &net.UDPAddr{IP: net.IPv4zero, Port: 0}
	m.enableIPv4(lAddr)
}

func (m *mClient) EnableIPv6() {
	var lAddr = &net.UDPAddr{IP: net.IPv6zero, Port: 0}
	m.enableIPv6(lAddr)
}
