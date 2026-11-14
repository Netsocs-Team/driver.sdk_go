package discovery

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/ipv4"
)

// InterfaceFilter picks the interfaces a scan goes out by. By default every
// interface that is up, has an IPv4 address and is not a loopback or a
// container bridge.
type InterfaceFilter struct {
	Only    []string // names to use, all others ignored
	Exclude []string // name prefixes to skip, on top of the defaults
}

// Environment of the netsocs-discovery service.
const (
	InterfacesEnv = "NETSOCS_DISCOVERY_INTERFACES" // comma-separated names
	ExcludeEnv    = "NETSOCS_DISCOVERY_EXCLUDE"    // comma-separated name prefixes
)

// FilterFromEnv reads InterfacesEnv and ExcludeEnv.
func FilterFromEnv() InterfaceFilter {
	split := func(s string) []string {
		var out []string
		for _, f := range strings.Split(s, ",") {
			if f = strings.TrimSpace(f); f != "" {
				out = append(out, f)
			}
		}
		return out
	}
	return InterfaceFilter{Only: split(os.Getenv(InterfacesEnv)), Exclude: split(os.Getenv(ExcludeEnv))}
}

var defaultExclude = []string{"docker", "br-", "veth", "virbr", "cni", "flannel", "cali", "lxc", "lo"}

type lanIface struct {
	iface net.Interface
	nets  []*net.IPNet
}

func (f InterfaceFilter) interfaces() ([]lanIface, error) {
	all, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var out []lanIface
	for _, ifc := range all {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		if len(f.Only) > 0 && !contains(f.Only, ifc.Name) {
			continue
		}
		if len(f.Only) == 0 && (hasPrefix(defaultExclude, ifc.Name) || hasPrefix(f.Exclude, ifc.Name)) {
			continue
		}
		addrs, _ := ifc.Addrs()
		var nets []*net.IPNet
		for _, a := range addrs {
			n, ok := a.(*net.IPNet)
			if !ok || n.IP.To4() == nil || n.IP.IsLinkLocalUnicast() || n.IP.IsLoopback() {
				continue
			}
			nets = append(nets, n)
		}
		if len(nets) > 0 {
			out = append(out, lanIface{ifc, nets})
		}
	}
	return out, nil
}

func contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func hasPrefix(prefixes []string, s string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func broadcastOf(n *net.IPNet) net.IP {
	ip, mask := n.IP.To4(), n.Mask
	if len(mask) == net.IPv6len {
		mask = mask[12:]
	}
	b := make(net.IP, 4)
	for i := range b {
		b[i] = ip[i] | ^mask[i]
	}
	return b
}

// One scan per answer port at a time: two scans on the same fixed port
// would steal each other's answers.
var (
	portLocksMu sync.Mutex
	portLocks   = map[int]*sync.Mutex{}
)

func lockPort(port int) func() {
	if port == 0 {
		return func() {}
	}
	portLocksMu.Lock()
	l, ok := portLocks[port]
	if !ok {
		l = &sync.Mutex{}
		portLocks[port] = l
	}
	portLocksMu.Unlock()
	l.Lock()
	return l.Unlock
}

// Execute does the scan on this host: it sends the payload out by every
// LAN interface and collects the answers until the timeout.
func Execute(ctx context.Context, p Probe, f InterfaceFilter) ([]Packet, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	ifaces, err := f.interfaces()
	if err != nil {
		return nil, fmt.Errorf("discovery: list interfaces: %w", err)
	}
	if len(ifaces) == 0 {
		return nil, errors.New("discovery: no LAN interface (up, with an IPv4 address) to scan from")
	}
	unlock := lockPort(p.ListenPort)
	defer unlock()

	lc := net.ListenConfig{Control: reuseAddr}
	c, err := lc.ListenPacket(ctx, "udp4", ":"+strconv.Itoa(p.ListenPort))
	if err != nil {
		return nil, fmt.Errorf("discovery: cannot listen on UDP %d (is another discovery tool using it?): %w", p.ListenPort, err)
	}
	defer c.Close()
	pc := ipv4.NewPacketConn(c)
	_ = pc.SetControlMessage(ipv4.FlagInterface, true) // not supported everywhere: best effort

	groups := map[string]bool{}
	for _, g := range p.Groups {
		groups[g] = true
	}
	if p.Dest != Broadcast {
		groups[p.Dest] = true
	}
	for g := range groups {
		for _, li := range ifaces {
			ifc := li.iface
			_ = pc.JoinGroup(&ifc, &net.UDPAddr{IP: net.ParseIP(g)}) // an interface without multicast just misses it
		}
	}

	var sent int
	var lastErr error
	send := func(dst net.IP, ifc *net.Interface) {
		if ifc != nil {
			if err := pc.SetMulticastInterface(ifc); err != nil {
				lastErr = err
				return
			}
		}
		if _, err := pc.WriteTo(p.Payload, nil, &net.UDPAddr{IP: dst, Port: p.DestPort}); err != nil {
			lastErr = err
			return
		}
		sent++
	}
	if p.Dest == Broadcast {
		seen := map[string]bool{}
		for _, li := range ifaces {
			for _, n := range li.nets {
				b := broadcastOf(n)
				if !seen[b.String()] {
					seen[b.String()] = true
					send(b, nil)
				}
			}
		}
		send(net.IPv4bcast, nil)
	} else {
		group := net.ParseIP(p.Dest)
		_ = pc.SetMulticastLoopback(false)
		for _, li := range ifaces {
			ifc := li.iface
			send(group, &ifc)
		}
	}
	if sent == 0 {
		return nil, fmt.Errorf("discovery: could not send the probe: %v", lastErr)
	}

	names := map[int]string{}
	for _, li := range ifaces {
		names[li.iface.Index] = li.iface.Name
	}
	deadline := time.Now().Add(p.Timeout)
	if d, ok := ctx.Deadline(); ok && d.Before(deadline) {
		deadline = d
	}
	var out []Packet
	buf := make([]byte, 65535)
	for len(out) < maxPackets {
		_ = pc.SetReadDeadline(deadline)
		n, cm, from, err := pc.ReadFrom(buf)
		if err != nil {
			break // deadline
		}
		data := buf[:n]
		if bytes.Equal(data, p.Payload) {
			continue // our own probe, looped back by a group or a broadcast
		}
		pkt := Packet{From: from.String(), Data: append([]byte(nil), data...)}
		if cm != nil {
			pkt.Interface = names[cm.IfIndex]
		}
		out = append(out, pkt)
	}
	return out, nil
}
