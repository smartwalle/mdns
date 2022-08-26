package main

import (
	"context"
	"fmt"
	"github.com/smartwalle/log4go"
	"github.com/smartwalle/mdns"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"time"
)

func main() {
	var name = mdns.MustName("xxxx.local.")
	_ = name
	var s = mdns.NewServer()
	s.EnableIPv4()
	//s.EnableIPv6()

	// Add all of our handlers
	s.OnQuestion(func(addr net.Addr, questions []mdns.Question) {
		for _, question := range questions {
			log4go.Println("OnQuestion", addr, question.Name, question.Type)
		}

		msg := dnsmessage.Message{
			Header: dnsmessage.Header{
				Response: true,
			},
			Answers: []dnsmessage.Resource{
				{
					Header: dnsmessage.ResourceHeader{
						Name:  name,
						Type:  dnsmessage.TypePTR,
						Class: dnsmessage.ClassINET,
					},
					Body: &dnsmessage.PTRResource{
						PTR: name,
					},
				},
				{
					Header: dnsmessage.ResourceHeader{
						Name:  name,
						Type:  dnsmessage.TypeA,
						Class: dnsmessage.ClassINET,
					},
					Body: &dnsmessage.AResource{A: mdns.IPv4ToBytes(net.ParseIP("192.168.1.99"))},
				},
				{
					Header: dnsmessage.ResourceHeader{
						Name:  name,
						Type:  dnsmessage.TypeAAAA,
						Class: dnsmessage.ClassINET,
					},
					Body: &dnsmessage.AAAAResource{AAAA: mdns.IPv6ToBytes(net.ParseIP("fe80::10ac:9ab5:ee60:9cfd"))},
				},
				{
					Header: dnsmessage.ResourceHeader{
						Name:  name,
						Type:  dnsmessage.TypeSRV,
						Class: dnsmessage.ClassINET,
					},
					Body: &dnsmessage.SRVResource{Port: 8000, Target: name},
				},
				{
					Header: dnsmessage.ResourceHeader{
						Name:  name,
						Type:  dnsmessage.TypeTXT,
						Class: dnsmessage.ClassINET,
					},
					Body: &dnsmessage.TXTResource{TXT: []string{"My awesome service 111"}},
				},
			},
		}

		fmt.Println(addr)
		s.SendTo(msg, addr.(*net.UDPAddr))
	})
	s.OnResource(func(addr net.Addr, resource mdns.Resource) {
		for _, answer := range resource.Answers {
			log4go.Println("OnResource", addr, answer.Header.Name, answer.Header.Type, answer.Body)
		}
	})

	s.OnWarning(func(addr net.Addr, err error) {
		log4go.Println("OnWarning", addr, err)
	})

	s.OnError(func(err error) {
		log4go.Println("OnError", err)
	})

	// Start up the mdns loop
	if err := s.Start(context.Background()); err != nil {
		log4go.Println("Start", err)
		return
	}

	//for {
	msg := dnsmessage.Message{
		Header: dnsmessage.Header{},
		Questions: []dnsmessage.Question{
			{
				Type:  dnsmessage.TypeA,
				Class: dnsmessage.ClassINET,
				Name:  name,
			},
		},
	}

	s.Multicast(msg)
	time.Sleep(time.Second * 1)
	fmt.Println("===========")
	//}

	// Now wait for either a timeout, an error, or an answer.
	select {}
}
