package main

import (
	"context"
	"github.com/smartwalle/log4go"
	"github.com/smartwalle/mdns"
	"golang.org/x/net/dns/dnsmessage"
	"net"
)

func main() {
	var name = mdns.MustName("smartwalle.local.")

	var server = mdns.NewServer()
	server.EnableIPv4()

	server.OnQuestion(func(addr net.Addr, question mdns.Question) {
		for _, q := range question.Questions {
			log4go.Println("OnQuestion", addr, q.Name, q.Type)

			if q.Name.String() == name.String() {
				var m = dnsmessage.Message{
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
							Body: &dnsmessage.TXTResource{TXT: []string{"My awesome service"}},
						},
					},
				}
				server.SendTo(m, addr.(*net.UDPAddr))
			}
		}
	})

	server.OnResource(func(addr net.Addr, resource mdns.Resource) {
		for _, answer := range resource.Answers {
			log4go.Println("OnResource", addr, answer.Header.Name, answer.Header.Type, answer.Body)
		}
	})

	server.OnWarning(func(addr net.Addr, err error) {
		log4go.Println("OnWarning", addr, err)
	})

	server.OnError(func(err error) {
		log4go.Println("OnError", err)
	})

	if err := server.Start(context.Background()); err != nil {
		log4go.Println("Start Error:", err)
		return
	}

	select {}
}
