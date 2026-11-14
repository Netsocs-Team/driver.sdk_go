// Package discovery lets a Netsocs driver find devices on the LAN.
//
// Device discovery protocols send a packet to a broadcast or multicast
// address and collect the answers. Drivers run inside the site container,
// on a Docker bridge network that broadcast and multicast never cross, so
// the scan itself is done by the netsocs-discovery service, which runs on
// the host network and is reached through a Unix socket. The driver keeps
// its protocol: it builds the request packet and parses the answers; the
// service only moves packets, and only those this package allows:
//
//   - to a broadcast or multicast address, never to a single host, so it
//     cannot be used to reach or scan arbitrary machines;
//   - to and from the known discovery ports of the supported brands;
//   - small payloads, for at most 10 seconds.
//
// When the socket is not there (a driver run by hand on a host), Run does
// the same scan itself.
package discovery

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

// Broadcast as Probe.Dest sends to 255.255.255.255 and to the broadcast
// address of every LAN interface.
const Broadcast = "broadcast"

// DefaultSocket is where the netsocs-discovery service listens.
const DefaultSocket = "/run/netsocs-discovery/probe.sock"

// SocketEnv overrides DefaultSocket.
const SocketEnv = "NETSOCS_DISCOVERY_SOCKET"

const (
	MaxTimeout = 10 * time.Second
	MaxPayload = 4096
	maxPackets = 1024
)

// Probe is one scan: Payload goes to Dest:DestPort from a socket bound to
// ListenPort (0: any free port, for protocols that answer the sender), which
// has joined Groups; the answers that arrive there within Timeout are
// returned.
type Probe struct {
	Dest       string        `json:"dest"` // multicast group (e.g. 239.255.255.251) or Broadcast
	DestPort   int           `json:"dest_port"`
	ListenPort int           `json:"listen_port"`
	Groups     []string      `json:"groups,omitempty"`
	Payload    []byte        `json:"payload"`
	Timeout    time.Duration `json:"timeout"`
}

// Packet is one answer.
type Packet struct {
	From      string `json:"from"`                // ip:port of the device
	Interface string `json:"interface,omitempty"` // local interface it came in by, when known
	Data      []byte `json:"data"`
}

// Allowed ports: the discovery protocols of the brands Netsocs integrates.
// Destination ports are where devices listen; listen ports are where they
// answer (0: the sending socket's own port).
var (
	allowedDestPorts = map[int]string{
		37810: "Dahua DHIP",
		5050:  "Dahua (old firmware)",
		7701:  "Hanwha IP Installer",
		37020: "Hikvision SADP",
		3702:  "ONVIF WS-Discovery",
	}
	allowedListenPorts = map[int]string{
		0:     "the sender's port",
		37810: "Dahua DHIP",
		5051:  "Dahua (old firmware)",
		7711:  "Hanwha IP Installer",
		37020: "Hikvision SADP",
	}
)

// AnswerPorts are the fixed UDP ports devices answer to: a host firewall
// must let them in (Windows Hyper-V firewall, ufw, firewalld).
func AnswerPorts() []int { return []int{37810, 5051, 7711, 37020} }

// Validate checks that the probe is one the service may send.
func (p *Probe) Validate() error {
	if p.Timeout <= 0 {
		p.Timeout = 3 * time.Second
	}
	if p.Timeout > MaxTimeout {
		return fmt.Errorf("discovery: timeout %s is over %s", p.Timeout, MaxTimeout)
	}
	if len(p.Payload) == 0 || len(p.Payload) > MaxPayload {
		return fmt.Errorf("discovery: payload must be 1 to %d bytes", MaxPayload)
	}
	if _, ok := allowedDestPorts[p.DestPort]; !ok {
		return fmt.Errorf("discovery: destination port %d is not a known discovery port", p.DestPort)
	}
	if _, ok := allowedListenPorts[p.ListenPort]; !ok {
		return fmt.Errorf("discovery: listen port %d is not a known discovery answer port", p.ListenPort)
	}
	if p.Dest != Broadcast {
		ip := net.ParseIP(p.Dest).To4()
		if ip == nil || !ip.IsMulticast() {
			return fmt.Errorf("discovery: destination %q must be %q or an IPv4 multicast group", p.Dest, Broadcast)
		}
	}
	if len(p.Groups) > 4 {
		return errors.New("discovery: at most 4 multicast groups")
	}
	for _, g := range p.Groups {
		ip := net.ParseIP(g).To4()
		if ip == nil || !ip.IsMulticast() {
			return fmt.Errorf("discovery: group %q is not an IPv4 multicast group", g)
		}
	}
	return nil
}

// Run sends the probe and returns the answers: through the
// netsocs-discovery service when its socket exists, by itself otherwise.
func Run(ctx context.Context, p Probe) ([]Packet, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	sock := os.Getenv(SocketEnv)
	if sock == "" {
		sock = DefaultSocket
	}
	if _, err := os.Stat(sock); err == nil {
		return runRemote(ctx, sock, p)
	}
	return Execute(ctx, p, InterfaceFilter{})
}
