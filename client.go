package mdns

import (
	"context"
	"github.com/smartwalle/mdns/internal"
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
	if m.conn4 == nil {
		var lAddr = &net.UDPAddr{
			IP:   net.IPv4zero,
			Port: 0,
		}
		var mAddr = &net.UDPAddr{
			IP:   mDNSMulticastIPv4,
			Port: Port,
		}
		m.conn4 = internal.NewConn(lAddr, mAddr, &internal.IPv4PacketConnFactory{}, -1)
	}
}

func (m *mClient) EnableIPv6() {
	if m.conn6 == nil {
		var lAddr = &net.UDPAddr{
			IP:   net.IPv6zero,
			Port: 0,
		}
		var mAddr = &net.UDPAddr{
			IP:   mDNSMulticastIPv6,
			Port: Port,
		}
		m.conn6 = internal.NewConn(lAddr, mAddr, &internal.IPv6PacketConnFactory{}, -1)
	}
}
