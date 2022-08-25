package mdns

import (
	"golang.org/x/net/dns/dnsmessage"
	"math/big"
	"net"
)

func MustName(name string) dnsmessage.Name {
	n, err := dnsmessage.NewName(name)
	if err != nil {
		panic(err)
	}
	return n
}

func IPv4ToBytes(ip net.IP) (out [4]byte) {
	rawIP := ip.To4()
	if rawIP == nil {
		return
	}

	ipInt := big.NewInt(0)
	ipInt.SetBytes(rawIP)
	copy(out[:], ipInt.Bytes())
	return
}

func IPv6ToBytes(ip net.IP) (out [16]byte) {
	rawIP := ip.To16()
	if rawIP == nil {
		return
	}

	ipInt := big.NewInt(0)
	ipInt.SetBytes(rawIP)
	copy(out[:], ipInt.Bytes())
	return
}

func IPToDNSRecordType(ip net.IP) dnsmessage.Type {
	if ip4 := ip.To4(); ip4 != nil {
		return dnsmessage.TypeA
	} else {
		return dnsmessage.TypeAAAA
	}
}
