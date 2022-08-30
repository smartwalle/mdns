package internal

import (
	"fmt"
	"golang.org/x/net/ipv4"
	"net"
)

type ipv4PacketConn struct {
	*ipv4.PacketConn
}

func (c *ipv4PacketConn) ReadFrom(b []byte) (int, net.Addr, error) {
	n, _, addr, err := c.PacketConn.ReadFrom(b)
	return n, addr, err
}

func (c *ipv4PacketConn) WriteTo(b []byte, dst net.Addr) (int, error) {
	return c.PacketConn.WriteTo(b, nil, dst)
}

type IPv4PacketConnFactory struct {
	Group *net.UDPAddr
}

func (f *IPv4PacketConnFactory) MakeUDPSocket(ifaces []net.Interface, addr *net.UDPAddr, ttl int) (net.PacketConn, error) {
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		return nil, err
	}

	pConn := ipv4.NewPacketConn(conn)
	if ttl >= 0 {
		if err := pConn.SetMulticastTTL(ttl); err != nil {
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
	return &ipv4PacketConn{PacketConn: pConn}, nil
}
