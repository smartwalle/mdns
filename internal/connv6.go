package internal

import (
	"golang.org/x/net/ipv6"
	"net"
)

type ipv6PacketConn struct {
	*ipv6.PacketConn
}

func (c *ipv6PacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, _, addr, err := c.PacketConn.ReadFrom(b)
	return n, addr, err
}

func (c *ipv6PacketConn) WriteTo(b []byte, dst net.Addr) (int, error) {
	return c.PacketConn.WriteTo(b, nil, dst)
}

type IPv6PacketConnFactory struct{}

func (*IPv6PacketConnFactory) MakeUDPSocket(ifaces []net.Interface, laddr *net.UDPAddr, ttl int) (net.PacketConn, error) {
	conn, err := net.ListenUDP("udp6", laddr)
	if err != nil {
		return nil, err
	}

	pConn := ipv6.NewPacketConn(conn)
	if ttl >= 0 {
		if err := pConn.SetMulticastHopLimit(ttl); err != nil {
			_ = pConn.Close()
			return nil, err
		}
	}

	for i, iface := range ifaces {
		if iface.Flags&net.FlagMulticast == 0 || iface.Flags&net.FlagPointToPoint == net.FlagPointToPoint {
			continue
		}
		if err := pConn.JoinGroup(&ifaces[i], laddr); err != nil {
			pConn.Close()
			return nil, err
		}
	}
	return &ipv6PacketConn{PacketConn: pConn}, nil
}

var defaultMDNSMulticastIPv6 = net.ParseIP("ff02::fb")

func NewIPv6Conn() *Conn {
	return NewConn(defaultMDNSMulticastIPv6, &IPv6PacketConnFactory{}, -1)
}
