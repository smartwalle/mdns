// Copyright 2019 The Fuchsia Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// This is a fork of https://fuchsia.googlesource.com/fuchsia/+/refs/heads/main/tools/net/mdns/mdns.go

package mdns

import (
	"context"
	"fmt"
	"github.com/smartwalle/mdns/internal"
	"golang.org/x/net/dns/dnsmessage"
	"net"
)

// Port is the mDNS port required of the spec
const Port = 5353

var mDNSMulticastIPv4 = net.ParseIP("224.0.0.251")
var mDNSMulticastIPv6 = net.ParseIP("ff02::fb")

var mDNSWildcardIPv4 = net.ParseIP("224.0.0.0")
var mDNSWildcardIPv6 = net.ParseIP("ff02::")

type mDNS struct {
	conn4    *internal.Conn
	conn6    *internal.Conn
	qHandler func(net.Addr, Question)
	rHandler func(net.Addr, Resource)
	wHandler func(net.Addr, error)
	eHandler func(error)
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

func (m *mDNS) OnQuestion(handler func(net.Addr, Question)) {
	m.qHandler = handler
}

func (m *mDNS) OnResource(handler func(net.Addr, Resource)) {
	m.rHandler = handler
}

// OnWarning calls f on every non-fatal error.
func (m *mDNS) OnWarning(handler func(net.Addr, error)) {
	m.wHandler = handler
}

// OnError calls f on every fatal error. After
// all active handlers are called, m will stop listening and
// close it's connection so this function will not be called twice.
func (m *mDNS) OnError(handler func(error)) {
	m.eHandler = handler
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

// Multicast serializes and sends packet out as a multicast to all interfaces
// using the port that m is listening on. Note that Start must be
// called prior to making this call.
func (m *mDNS) Multicast(message dnsmessage.Message) error {
	var b, err = message.Pack()
	if err != nil {
		return err
	}

	var err4 error
	if m.conn4 != nil {
		err4 = m.conn4.Multicast(b)
	}
	var err6 error
	if m.conn6 != nil {
		err6 = m.conn6.Multicast(b)
	}
	if err4 != nil {
		return err4
	}
	return err6
}

func (m *mDNS) initMDNSConn() error {
	if m.conn4 == nil && m.conn6 == nil {
		return fmt.Errorf("no connection active")
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("listing interfaces: %w", err)
	}

	if c := m.conn4; c != nil {
		if err = c.MakeUDPSocket(ifaces); err != nil {
			return err
		}
	}
	if c := m.conn6; c != nil {
		if err = c.MakeUDPSocket(ifaces); err != nil {
			return err
		}
	}
	return nil
}

func (m *mDNS) Start(ctx context.Context) error {
	if err := m.initMDNSConn(); err != nil {
		m.Close()
		return err
	}
	go func() {
		// NOTE: This defer statement will close connections, which will force
		// the goroutines started by Listen() to exit.
		defer m.Close()

		var quit = make(chan struct{})
		defer close(quit)

		var packets = make(chan internal.Packet, 1)

		if m.conn4 != nil {
			m.conn4.Listen(packets, quit)
		}
		if m.conn6 != nil {
			m.conn6.Listen(packets, quit)
		}

		var parser = &dnsmessage.Parser{}

		for {
			var received internal.Packet

			select {
			case <-ctx.Done():
				return
			case received = <-packets:
			}

			if received.Error != nil {
				if m.eHandler != nil {
					go m.eHandler(received.Error)
				}
				return
			}

			var header dnsmessage.Header
			var err error

			if header, err = parser.Start(received.Data); err != nil {
				if m.wHandler != nil {
					go m.wHandler(received.Addr, err)
				}
				continue
			}

			if m.qHandler != nil {
				var questions, _ = parser.AllQuestions()
				if len(questions) > 0 {
					var nQuestion = Question{
						Header:    header,
						Questions: questions,
					}
					go m.qHandler(received.Addr, nQuestion)
				}
			} else {
				parser.SkipAllQuestions()
			}

			if m.rHandler != nil {
				var answers, _ = parser.AllAnswers()
				var authorities, _ = parser.AllAuthorities()
				var additionals, _ = parser.AllAdditionals()

				if len(answers) > 0 || len(authorities) > 0 || len(additionals) > 0 {
					var nResource = Resource{
						Header:      header,
						Answers:     answers,
						Authorities: authorities,
						Additionals: additionals,
					}
					go m.rHandler(received.Addr, nResource)
				}
			} else {
				parser.SkipAllAnswers()
				parser.SkipAllAuthorities()
				parser.SkipAllAdditionals()
			}
		}
	}()
	return nil
}
