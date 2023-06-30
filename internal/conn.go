package internal

import (
	"fmt"
	"net"
	"os"
	"syscall"
)

type PacketConnFactory interface {
	MakeUDPSocket(iface []net.Interface, addr *net.UDPAddr, ttl int) (net.PacketConn, error)
}

func NewConn(mAddr, lAddr, rAddr *net.UDPAddr, lFactory, rFactory PacketConnFactory, ttl int) *Conn {
	return &Conn{
		mAddr:    mAddr,
		lAddr:    lAddr,
		rAddr:    rAddr,
		lFactory: lFactory,
		rFactory: rFactory,
		ttl:      ttl,
	}
}

type Conn struct {
	mAddr    *net.UDPAddr
	lAddr    *net.UDPAddr
	rAddr    *net.UDPAddr
	lFactory PacketConnFactory
	rFactory PacketConnFactory
	lConn    net.PacketConn
	rConn    net.PacketConn
	ttl      int
}

func (c *Conn) SetMulticastTTL(ttl int) error {
	if ttl > 255 {
		return fmt.Errorf("TTL outside of valid range: %d", ttl)
	}
	c.ttl = ttl
	return nil
}

func (c *Conn) SetPort(mPort, lPort, rPort int) {
	if c.mAddr != nil {
		c.mAddr.Port = mPort
	}
	if c.lAddr != nil {
		c.lAddr.Port = lPort
	}
	if c.rAddr != nil {
		c.rAddr.Port = rPort
	}
}

func (c *Conn) Close() error {
	if c.lConn != nil {
		c.lConn.Close()
		c.lConn = nil
	}
	if c.rConn != nil {
		c.rConn.Close()
		c.rConn = nil
	}
	return nil
}

func (c *Conn) Listen(packets chan Packet, quit chan struct{}) {
	go c.listen(c.lConn, packets, quit)

	if c.rConn != nil {
		go c.listen(c.rConn, packets, quit)
	}
}

func (c *Conn) listen(conn net.PacketConn, packets chan Packet, quit chan struct{}) {
	var payload = make([]byte, 1<<16)
	for {
		n, src, err := conn.ReadFrom(payload)
		if err != nil {
			select {
			case <-quit:
			case packets <- Packet{Error: err}:
			}
			return
		}
		select {
		case <-quit:
		case packets <- Packet{Data: append([]byte(nil), payload[:n]...), Addr: src}:
		}
	}
}

func (c *Conn) SendTo(b []byte, dst *net.UDPAddr) error {
	_, err := c.lConn.WriteTo(b, dst)
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

func (c *Conn) Multicast(b []byte) error {
	return c.SendTo(b, c.mAddr)
}

func (c *Conn) MakeUDPSocket(ifaces []net.Interface) (err error) {
	var lConn net.PacketConn
	var rConn net.PacketConn

	lConn, err = c.lFactory.MakeUDPSocket(ifaces, c.lAddr, c.ttl)
	if err != nil {
		return err
	}

	if c.rFactory != nil && c.rAddr != nil {
		rConn, err = c.rFactory.MakeUDPSocket(ifaces, c.rAddr, c.ttl)
		if err != nil {
			lConn.Close()
			return err
		}
	}

	c.lConn = lConn
	c.rConn = rConn
	return nil
}
