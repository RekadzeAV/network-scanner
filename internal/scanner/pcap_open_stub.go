//go:build !((windows && (386 || amd64)) || (unix && cgo))

package scanner

import "errors"

// openLivePcap is unavailable in builds where gopacket/pcap cannot be
// compiled (e.g. cross-compiled unix targets with CGO_ENABLED=0,
// windows/arm64 without cgo): gopacket/pcap requires libpcap/npcap
// and a cgo toolchain.
func openLivePcap(device string, snaplen int32, promiscuous bool) (livePacketHandle, error) {
	return nil, errors.New("pcap not available in this build (CGO_ENABLED=0 or unsupported target)")
}
