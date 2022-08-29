package main

import (
	"context"
	"github.com/smartwalle/log4go"
	"github.com/smartwalle/mdns"
	"golang.org/x/net/dns/dnsmessage"
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
		log4go.Println(addr)
		for _, answer := range resource.Answers {
			if answer.Header.Name.String() != name.String() {
				continue
			}
			found <- Resource{Resource: resource, Addr: addr}
		}
	})

	client.OnWarning(func(addr net.Addr, err error) {
		log4go.Println("OnWarning", addr, err)
	})

	client.OnError(func(err error) {
		log4go.Println("OnError", err)
		os.Exit(-1)
	})

	if err := client.Start(context.Background()); err != nil {
		log4go.Println("Start Error:", err)
		return
	}

	var m = dnsmessage.Message{
		Header: dnsmessage.Header{},
		Questions: []dnsmessage.Question{
			{
				Type:  dnsmessage.TypeA,
				Class: dnsmessage.ClassINET,
				Name:  name,
			},
		},
	}
	if err := client.Multicast(m); err != nil {
		log4go.Println("Multicast Error:", err)
		return
	}

	select {
	case <-time.After(time.Second * 30):
		log4go.Println("Timeout")
	case resource := <-found:
		for _, answer := range resource.Answers {
			log4go.Println(resource.Addr, answer.Header.Name, answer.Header.Type, answer.Body)
		}
	}
}
