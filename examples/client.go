package main

import (
	"context"
	"github.com/smartwalle/mdns"
	"golang.org/x/net/dns/dnsmessage"
	"log/slog"
	"net"
	"os"
	"time"
)

type Resource struct {
	mdns.Resource
	Addr net.Addr
}

func main() {
	var name = mdns.MustName("smartwalle.local.")

	var client = mdns.NewClient()
	client.EnableIPv4()

	var found = make(chan Resource, 1)
	client.OnResource(func(addr net.Addr, resource mdns.Resource) {
		slog.Info("OnResource", slog.Any("addr", addr))
		for _, answer := range resource.Answers {
			if answer.Header.Name.String() != name.String() {
				continue
			}
			found <- Resource{Resource: resource, Addr: addr}
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
		slog.Info("Multicast Error", slog.Any("error", err))
		return
	}

	select {
	case <-time.After(time.Second * 30):
		slog.Info("Timeout")
	case resource := <-found:
		for _, answer := range resource.Answers {
			slog.Info("Answer", slog.Any("addr", resource.Addr), slog.Any("name", answer.Header.Name), slog.Any("type", answer.Header.Type), slog.Any("body", answer.Body))
		}
	}
}
