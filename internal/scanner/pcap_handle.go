package scanner

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

// livePacketHandle abstracts *pcap.Handle so that the ARP probe code compiles
// on platforms/builds where cgo (and thus libpcap) is unavailable.
type livePacketHandle interface {
	gopacket.PacketDataSource
	WritePacketData(data []byte) error
	LinkType() layers.LinkType
	Close()
}
