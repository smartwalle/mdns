package main

import (
	"context"
	"fmt"
	"github.com/smartwalle/log4go"
	"github.com/smartwalle/mdns"
	"golang.org/x/net/dns/dnsmessage"
	"net"
)

func main() {
	var name = mdns.MustName("xxxx.local.")
	_ = name

	var c = mdns.NewClient()
	c.EnableIPv4()

	c.OnResource(func(addr net.Addr, resource mdns.Resource) {
		for _, answer := range resource.Answers {
			log4go.Println("OnResource", addr, answer.Header.Name, answer.Header.Type, answer.Body)
		}
	})

	c.Start(context.Background())

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

	fmt.Println(c.Multicast(msg))

	select {}
}
