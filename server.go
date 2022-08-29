package mdns

import (
	"context"
	"github.com/smartwalle/mdns/internal"
	"golang.org/x/net/dns/dnsmessage"
	"net"
)

// Server is the central interface through which requests are sent and received.
// This implementation is agnostic to use case and asynchronous.
// To handle various responses add Handlers. To send a packet you may use
// either SendTo (generally used for unicast) or Send (generally used for
// multicast).
type Server interface {
	// EnableIPv4 enables listening on IPv4 network interfaces.
	EnableIPv4()

	// EnableIPv6 enables listening on IPv6 network interfaces.
	EnableIPv6()

	// SetMulticastTTL sets the multicast time to live. If this is set to less
	// than zero it stays at the default. If it is set to zero this will mean
	// no packets can escape the host.
	//
	// Must be no greater than 255.
	SetMulticastTTL(ttl int) error

	// OnQuestion calls f on every Question received.
	OnQuestion(f func(net.Addr, Question))

	// OnResource calls f on every Resource received.
	OnResource(f func(net.Addr, Resource))

	// OnWarning calls f on every non-fatal error.
	OnWarning(f func(net.Addr, error))

	// OnError calls f on every fatal error. After
	// all active handlers are called, m will stop listening and
	// close it's connection so this function will not be called twice.
	OnError(f func(error))

	// Start causes m to start listening for mDNS packets on all interfaces on
	// the specified port. Listening will stop if ctx is done.
	Start(ctx context.Context) error

	// SendTo serializes and sends packet to dst. If dst is a multicast
	// address then packet is multicast to the corresponding group on
	// all interfaces. Note that start must be called prior to making this
	// call.
	SendTo(resource Resource, dst *net.UDPAddr) error

	// Multicast serializes and sends packet out as a multicast to all interfaces
	// using the port that m is listening on. Note that Start must be
	// called prior to making this call.
	Multicast(resource Resource) error

	// Close closes all connections.
	Close()
}

type mServer struct {
	*mDNS
}

// NewServer creates a new object implementing the Server interface. Do not forget
// to call EnableIPv4() or EnableIPv6() to enable listening on interfaces of
// the corresponding type, or nothing will work.
func NewServer() Server {
	var nServer = &mServer{}
	nServer.mDNS = &mDNS{}
	nServer.mDNS.conn4 = nil
	nServer.mDNS.conn6 = nil
	return nServer
}

func (m *mServer) EnableIPv4() {
	if m.conn4 == nil {
		var mAddr = &net.UDPAddr{
			IP:   mDNSMulticastIPv4,
			Port: Port,
		}
		var lAddr = &net.UDPAddr{
			IP:   mDNSWildcardIPv4,
			Port: Port,
		}
		m.conn4 = internal.NewConn(
			mAddr,
			lAddr,
			nil,
			&internal.IPv4PacketConnFactory{Group: &net.UDPAddr{IP: mDNSMulticastIPv4}},
			nil,
			-1,
		)
	}
}

func (m *mServer) EnableIPv6() {
	if m.conn6 == nil {
		var mAddr = &net.UDPAddr{
			IP:   mDNSMulticastIPv6,
			Port: Port,
		}
		var lAddr = &net.UDPAddr{
			IP:   mDNSWildcardIPv6,
			Port: Port,
		}
		m.conn6 = internal.NewConn(
			mAddr,
			lAddr,
			nil,
			&internal.IPv6PacketConnFactory{Group: &net.UDPAddr{IP: mDNSMulticastIPv6}},
			nil,
			-1,
		)
	}
}

func (m *mServer) SendTo(resource Resource, dst *net.UDPAddr) error {
	var message = dnsmessage.Message{
		Header:      resource.Header,
		Answers:     resource.Answers,
		Authorities: resource.Authorities,
		Additionals: resource.Additionals,
	}
	return m.mDNS.SendTo(message, dst)
}

func (m *mServer) Multicast(resource Resource) error {
	var message = dnsmessage.Message{
		Header:      resource.Header,
		Answers:     resource.Answers,
		Authorities: resource.Authorities,
		Additionals: resource.Additionals,
	}
	return m.mDNS.Multicast(message)
}
