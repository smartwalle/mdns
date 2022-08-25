// Copyright 2019 The Fuchsia Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This is a fork of https://fuchsia.googlesource.com/fuchsia/+/refs/heads/main/tools/net/mdns/mdns.go written in a
// more extendable way

package mdns

import (
	"context"
	"fmt"
	"github.com/smartwalle/mdns/internal"
	"golang.org/x/net/dns/dnsmessage"
	"net"
)

// DefaultPort is the mDNS port required of the spec, though this library is port-agnostic.
const DefaultPort int = 5353

// MDNS is the central interface through which requests are sent and received.
// This implementation is agnostic to use case and asynchronous.
// To handle various responses add Handlers. To send a packet you may use
// either SendTo (generally used for unicast) or Send (generally used for
// multicast).
type MDNS interface {
	// EnableIPv4 enables listening on IPv4 network interfaces.
	EnableIPv4()

	// EnableIPv6 enables listening on IPv6 network interfaces.
	EnableIPv6()

	// SetAddress sets a non-default listen address.
	SetAddress(address string) error

	// SetMulticastTTL sets the multicast time to live. If this is set to less
	// than zero it stays at the default. If it is set to zero this will mean
	// no packets can escape the host.
	//
	// Must be no greater than 255.
	SetMulticastTTL(ttl int) error

	// OnQuestion calls f on every Question received.
	OnQuestion(f func(net.Addr, dnsmessage.Question))

	// OnResource calls f on every Resource received.
	OnResource(f func(net.Addr, dnsmessage.Resource))

	// OnWarning calls f on every non-fatal error.
	OnWarning(f func(net.Addr, error))

	// OnError calls f on every fatal error. After
	// all active handlers are called, m will stop listening and
	// close it's connection so this function will not be called twice.
	OnError(f func(error))

	// Start causes m to start listening for mDNS packets on all interfaces on
	// the specified port. Listening will stop if ctx is done.
	Start(ctx context.Context, port int) error

	// Send serializes and sends packet out as a multicast to all interfaces
	// using the port that m is listening on. Note that Start must be
	// called prior to making this call.
	Send(message dnsmessage.Message) error

	// SendTo serializes and sends packet to dst. If dst is a multicast
	// address then packet is multicast to the corresponding group on
	// all interfaces. Note that start must be called prior to making this
	// call.
	SendTo(message dnsmessage.Message, dst *net.UDPAddr) error

	// Close closes all connections.
	Close()
}

type mDNS struct {
	conn4    *internal.Conn
	conn6    *internal.Conn
	port     int
	qHandler func(net.Addr, dnsmessage.Question)
	rHandler func(net.Addr, dnsmessage.Resource)
	wHandler func(net.Addr, error)
	eHandler func(error)
}

// New creates a new object implementing the MDNS interface. Do not forget
// to call EnableIPv4() or EnableIPv6() to enable listening on interfaces of
// the corresponding type, or nothing will work.
func New() MDNS {
	m := mDNS{}
	m.conn4 = nil
	m.conn6 = nil
	return &m
}

func (m *mDNS) EnableIPv4() {
	if m.conn4 == nil {
		m.conn4 = internal.NewIPv4Conn()
	}
}

func (m *mDNS) EnableIPv6() {
	if m.conn6 == nil {
		m.conn6 = internal.NewIPv6Conn()
	}
}

func (m *mDNS) Close() {
	if m.conn4 != nil {
		m.conn4.Close()
		m.conn4 = nil
	}
	if m.conn6 != nil {
		m.conn6.Close()
		m.conn6 = nil
	}
}

func (m *mDNS) SetAddress(address string) error {
	ip := net.ParseIP(address)
	if ip4 := ip.To4(); ip4 != nil {
		if m.conn4 == nil {
			return fmt.Errorf("mDNS IPv4 support is disabled")
		}
		m.conn4.SetIP(ip4)
	} else if ip16 := ip.To16(); ip16 != nil {
		if m.conn6 == nil {
			return fmt.Errorf("mDNS IPv6 support is disabled")
		}
		m.conn4.SetIP(ip16)
	} else {
		return fmt.Errorf("not a valid IP address: %s", address)
	}
	return nil
}

func (m *mDNS) SetMulticastTTL(ttl int) error {
	if m.conn4 != nil {
		if err := m.conn4.SetMulticastTTL(ttl); err != nil {
			return err
		}
	}
	if m.conn6 != nil {
		if err := m.conn6.SetMulticastTTL(ttl); err != nil {
			return err
		}
	}
	return nil
}

func (m *mDNS) OnQuestion(f func(net.Addr, dnsmessage.Question)) {
	m.qHandler = f
}

func (m *mDNS) OnResource(f func(net.Addr, dnsmessage.Resource)) {
	m.rHandler = f
}

// OnWarning calls f on every non-fatal error.
func (m *mDNS) OnWarning(f func(net.Addr, error)) {
	m.wHandler = f
}

// OnError calls f on every fatal error. After
// all active handlers are called, m will stop listening and
// close it's connection so this function will not be called twice.
func (m *mDNS) OnError(f func(error)) {
	m.eHandler = f
}

// Send serializes and sends packet out as a multicast to all interfaces
// using the port that m is listening on. Note that Start must be
// called prior to making this call.
func (m *mDNS) Send(message dnsmessage.Message) error {
	var b, err = message.Pack()
	if err != nil {
		return err
	}

	var err4 error
	if m.conn4 != nil {
		err4 = m.conn4.Send(b)
	}
	var err6 error
	if m.conn6 != nil {
		err6 = m.conn6.Send(b)
	}
	if err4 != nil {
		return err4
	}
	return err6
}

// SendTo serializes and sends packet to dst. If dst is a multicast
// address then packet is multicast to the corresponding group on
// all interfaces. Note that start must be called prior to making this
// call.
func (m *mDNS) SendTo(message dnsmessage.Message, dst *net.UDPAddr) error {
	var b, err = message.Pack()
	if err != nil {
		return err
	}

	if dst.IP.To4() != nil {
		if m.conn4 != nil {
			return m.conn4.SendTo(b, dst)
		} else {
			return fmt.Errorf("IPv4 was not enabled!")
		}
	} else {
		if m.conn6 != nil {
			return m.conn6.SendTo(b, dst)
		} else {
			return fmt.Errorf("IPv6 was not enabled!")
		}
	}
}

func (m *mDNS) initMDNSConn(port int) error {
	if m.conn4 == nil && m.conn6 == nil {
		return fmt.Errorf("no connection active")
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("listing interfaces: %w", err)
	}

	if c := m.conn4; c != nil {
		if err = c.MakeUDPSocket(ifaces, port); err != nil {
			return err
		}
	}
	if c := m.conn6; c != nil {
		if err = c.MakeUDPSocket(ifaces, port); err != nil {
			return err
		}
	}
	return nil
}

// Start causes m to start listening for MDNS packets on all interfaces on
// the specified port. Listening will stop if ctx is done.
func (m *mDNS) Start(ctx context.Context, port int) error {
	if err := m.initMDNSConn(port); err != nil {
		m.Close()
		return err
	}
	go func() {
		// NOTE: This defer statement will close connections, which will force
		// the goroutines started by Listen() to exit.
		defer m.Close()

		var chan4 <-chan internal.Packet
		var chan6 <-chan internal.Packet

		if m.conn4 != nil {
			chan4 = m.conn4.Listen()
		}
		if m.conn6 != nil {
			chan6 = m.conn6.Listen()
		}

		var p = dnsmessage.Parser{}

		for {
			var received internal.Packet

			select {
			case <-ctx.Done():
				return
			case received = <-chan4:
				break
			case received = <-chan6:
				break
			}

			if received.Error != nil {
				if m.eHandler != nil {
					go m.eHandler(received.Error)
				}
				return
			}

			if _, err := p.Start(received.Data); err != nil {
				if m.wHandler != nil {
					go m.wHandler(received.Addr, err)
				}
				continue
			}

			if m.qHandler != nil {
				var questions, _ = p.AllQuestions()
				for _, question := range questions {
					go m.qHandler(received.Addr, question)
				}
			} else {
				p.SkipAllQuestions()
			}

			if m.rHandler != nil {
				var answers, _ = p.AllAnswers()
				for _, answer := range answers {
					go m.rHandler(received.Addr, answer)
				}
			} else {
				p.SkipAllAnswers()
			}
		}
	}()
	return nil
}
