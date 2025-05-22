package main

import (
	"context"
	"github.com/smartwalle/mdns"
	"golang.org/x/net/dns/dnsmessage"
	"log/slog"
	"net"
	"os"
)

func main() {
	var name = mdns.MustName("smartwalle.local.")

	var client = mdns.NewClient()
	client.EnableIPv4()

	client.OnResource(func(addr net.Addr, resource mdns.Resource) {
		slog.Info("----------- OnResource", slog.Any("header", resource.Header))
		for _, answer := range resource.Answers {
			if answer.Header.Name.String() != name.String() {
				continue
			}

			slog.Info("OnResource", slog.Any("addr", addr), slog.Any("name", answer.Header.Name), slog.Any("type", answer.Header.Type), slog.Any("body", answer.Body))
		}
	})

	client.OnWarning(func(addr net.Addr, err error) {
		slog.Info("OnWarning", slog.Any("addr", addr), slog.Any("error", err))
	})

	client.OnError(func(err error) {
		slog.Info("OnError", slog.Any("error", err))
		os.Exit(-1)
	})

	if err := client.Start(context.Background()); err != nil {
		slog.Info("Start Error", slog.Any("error", err))
		os.Exit(-1)
		return
	}

	var question = mdns.Question{
		Header: dnsmessage.Header{},
		Questions: []dnsmessage.Question{
			{
				Type:  dnsmessage.TypeA,
				Class: dnsmessage.ClassINET,
				Name:  name,
			},
		},
	}
	if err := client.Send(question); err != nil {
		slog.Info("Send Error", slog.Any("error", err))
		return
	}

	select {}
}
