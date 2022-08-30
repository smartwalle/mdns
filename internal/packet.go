package internal

import "net"

// Packet A small struct used to send received UDP packets and
// information about their interface / source address through a channel.
type Packet struct {
	Addr  net.Addr
	Error error
	Data  []byte
}
