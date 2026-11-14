package discovery

import (
	"net"
	"testing"
	"time"
)

func TestValidateAllowsOnlyBroadcastOrMulticastToKnownPorts(t *testing.T) {
	ok := []Probe{
		{Dest: "239.255.255.251", DestPort: 37810, ListenPort: 37810, Payload: []byte("x")},
		{Dest: Broadcast, DestPort: 5050, ListenPort: 5051, Payload: []byte("x")},
		{Dest: Broadcast, DestPort: 7701, ListenPort: 7711, Payload: []byte("x")},
		{Dest: "239.255.255.250", DestPort: 37020, ListenPort: 37020, Payload: []byte("x")},
		{Dest: "239.255.255.250", DestPort: 3702, ListenPort: 0, Payload: []byte("x")},
		{Dest: "239.255.255.250", DestPort: 3702, ListenPort: 3703, Payload: []byte("x")},
	}
	for _, p := range ok {
		if err := p.Validate(); err != nil {
			t.Errorf("%+v: %v", p, err)
		}
	}
	bad := map[string]Probe{
		"unicast host":      {Dest: "192.168.0.14", DestPort: 37810, Payload: []byte("x")},
		"hostname":          {Dest: "camera.local", DestPort: 37810, Payload: []byte("x")},
		"other port":        {Dest: Broadcast, DestPort: 22, Payload: []byte("x")},
		"other listen port": {Dest: Broadcast, DestPort: 7701, ListenPort: 8080, Payload: []byte("x")},
		"empty payload":     {Dest: Broadcast, DestPort: 7701},
		"big payload":       {Dest: Broadcast, DestPort: 7701, Payload: make([]byte, MaxPayload+1)},
		"long timeout":      {Dest: Broadcast, DestPort: 7701, Payload: []byte("x"), Timeout: time.Minute},
		"unicast group":     {Dest: Broadcast, DestPort: 7701, Payload: []byte("x"), Groups: []string{"10.0.0.1"}},
	}
	for name, p := range bad {
		if err := p.Validate(); err == nil {
			t.Errorf("%s must be refused", name)
		}
	}
}

func TestBroadcastOf(t *testing.T) {
	_, n, _ := net.ParseCIDR("192.168.0.22/24")
	n.IP = net.ParseIP("192.168.0.22").To4()
	if b := broadcastOf(n); b.String() != "192.168.0.255" {
		t.Fatal(b)
	}
	_, n, _ = net.ParseCIDR("10.214.40.56/16")
	n.IP = net.ParseIP("10.214.40.56").To4()
	if b := broadcastOf(n); b.String() != "10.214.255.255" {
		t.Fatal(b)
	}
}

func TestInterfaceFilterSkipsContainerBridges(t *testing.T) {
	f := InterfaceFilter{}
	for _, name := range []string{"docker0", "br-1234", "veth9a", "lo"} {
		if !hasPrefix(defaultExclude, name) {
			t.Errorf("%s must be excluded by default", name)
		}
	}
	for _, name := range []string{"eth0", "eth3", "enp3s0", "wlan0"} {
		if hasPrefix(append(defaultExclude, f.Exclude...), name) {
			t.Errorf("%s must be scanned", name)
		}
	}
}
