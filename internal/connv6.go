package internal

import (
	"fmt"
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

type IPv6PacketConnFactory struct {
	Group *net.UDPAddr
}

func (f *IPv6PacketConnFactory) MakeUDPSocket(ifaces []net.Interface, addr *net.UDPAddr, ttl int) (net.PacketConn, error) {
	conn, err := net.ListenUDP("udp6", addr)
	if err != nil {
		return nil, err
	}

	pConn := ipv6.NewPacketConn(conn)
	if ttl >= 0 {
		if err := pConn.SetMulticastHopLimit(ttl); err != nil {
			pConn.Close()
			return nil, err
		}
	}

	if f.Group != nil {
		pConn.SetMulticastLoopback(true)

		var errCount = 0

		for i := range ifaces {
			if nErr := pConn.JoinGroup(&ifaces[i], f.Group); nErr != nil {
				errCount += 1
			}
		}

		if errCount == len(ifaces) {
			pConn.Close()
			return nil, fmt.Errorf("failed to join multicast group on all interfaces")
		}
	}
	return &ipv6PacketConn{PacketConn: pConn}, nil
}
