//go:build (windows && (386 || amd64)) || (unix && cgo)

package scanner

import (
	"github.com/google/gopacket/pcap"
)

// openLivePcap opens a network interface for raw packet I/O via libpcap
// (npcap on Windows). Available where gopacket/pcap compiles.
func openLivePcap(device string, snaplen int32, promiscuous bool) (livePacketHandle, error) {
	return pcap.OpenLive(device, snaplen, promiscuous, pcap.BlockForever)
}
