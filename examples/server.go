package main

import (
	"context"
	"github.com/smartwalle/mdns"
	"golang.org/x/net/dns/dnsmessage"
	"log/slog"
	"net"
)

func main() {
	var name = mdns.MustName("smartwalle.local.")

	var server = mdns.NewServer()
	server.EnableIPv4()

	server.OnQuestion(func(addr net.Addr, question mdns.Question) {
		slog.Info("----------- OnQuestion", slog.Any("header", question.Header))
		for _, q := range question.Questions {
			slog.Info("OnQuestion", slog.Any("addr", addr), slog.Any("name", q.Name), slog.Any("type", q.Type))

			if q.Name.String() == name.String() {
				var resource = mdns.Resource{
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
				if err := server.SendTo(resource, addr.(*net.UDPAddr)); err != nil {
					slog.Info("Send Error", slog.Any("error", err))
				}
			}
		}
	})

	server.OnResource(func(addr net.Addr, resource mdns.Resource) {
		slog.Info("----------- OnResource", slog.Any("header", resource.Header))
		for _, answer := range resource.Answers {
			//if answer.Header.Name == name {
			slog.Info("OnResource", slog.Any("addr", addr), slog.Any("name", answer.Header.Name), slog.Any("type", answer.Header.Type), slog.Any("body", answer.Body))
			//}
		}
	})

	server.OnWarning(func(addr net.Addr, err error) {
		slog.Info("OnWarning", slog.Any("addr", addr), slog.Any("error", err))
	})

	server.OnError(func(err error) {
		slog.Info("OnError", slog.Any("error", err))
	})

	if err := server.Start(context.Background()); err != nil {
		slog.Info("Start Error", slog.Any("error", err))
		return
	}

	var resource = mdns.Resource{
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

	if err := server.Multicast(resource); err != nil {
		slog.Info("Multicast Error", slog.Any("error", err))
		return
	}

	select {}
}
