package internal

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

type PacketConnFactory interface {
	MakeUDPSocket(iface []net.Interface, laddr *net.UDPAddr, ttl int) (net.PacketConn, error)
}

func NewConn(ip net.IP, factory PacketConnFactory, ttl int) *Conn {
	return &Conn{
		dst: net.UDPAddr{
			IP: ip,
		},
		factory: factory,
		ttl:     ttl,
	}
}

type Conn struct {
	dst     net.UDPAddr
	ttl     int
	factory PacketConnFactory
	conn    net.PacketConn
}

func (c *Conn) SetIP(ip net.IP) {
	c.dst.IP = ip
}

func (c *Conn) SetMulticastTTL(ttl int) error {
	if ttl > 255 {
		return fmt.Errorf("TTL outside of valid range: %d", ttl)
	}
	c.ttl = ttl
	return nil
}

func (c *Conn) Close() error {
	if c.conn == nil {
		return nil
	}
	if err := c.conn.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Conn) Listen() <-chan Packet {
	ch := make(chan Packet, 1)

	go func(conn net.PacketConn) {
		payload := make([]byte, 1<<16)
		for {
			n, src, err := conn.ReadFrom(payload)
			if err != nil {
				ch <- Packet{Error: err}
				return
			}
			ch <- Packet{
				Data: append([]byte(nil), payload[:n]...),
				Addr: src,
			}
		}
	}(c.conn)
	return ch
}

func (c *Conn) SendTo(b []byte, dst *net.UDPAddr) error {
	_, err := c.conn.WriteTo(b, dst)
	if err != nil {
		if err, ok := err.(*net.OpError); ok {
			if err, ok := err.Err.(*os.SyscallError); ok {
				if err, ok := err.Err.(syscall.Errno); ok {
					switch err {
					case syscall.EADDRNOTAVAIL, syscall.ENETUNREACH:
					}
				}
			}
		}
		return err
	}
	return nil
}

func (c *Conn) Send(b []byte) error {
	return c.SendTo(b, &c.dst)
}

func (c *Conn) MakeUDPSocket(ifaces []net.Interface, port int) error {
	c.dst.Port = port
	conn, err := c.factory.MakeUDPSocket(ifaces, &c.dst, c.ttl)
	if err != nil {
		return err
	}
	c.conn = conn
	return nil
}
