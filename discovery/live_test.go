package discovery

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// dhipSearch is a Dahua DHIP search, used as a real probe on the LAN.
func dhipSearch() []byte {
	body := []byte(`{"method":"DHDiscover.search","params":{"mac":"","uni":1}}`)
	h := make([]byte, 32)
	binary.LittleEndian.PutUint32(h[0:], 32)
	copy(h[4:], "DHIP")
	binary.LittleEndian.PutUint32(h[16:], uint32(len(body)))
	binary.LittleEndian.PutUint32(h[24:], uint32(len(body)))
	return append(h, body...)
}

// Run with NETSOCS_DISCOVERY_LIVE=1 on a LAN with Dahua devices: scans
// directly and through a relay on a temporary socket.
func TestLive_DHIPDirectAndThroughRelay(t *testing.T) {
	if os.Getenv("NETSOCS_DISCOVERY_LIVE") == "" {
		t.Skip("set NETSOCS_DISCOVERY_LIVE=1")
	}
	p := Probe{Dest: "239.255.255.251", DestPort: 37810, ListenPort: 37810, Payload: dhipSearch(), Timeout: 3 * time.Second}
	direct, err := Execute(context.Background(), p, FilterFromEnv())
	if err != nil {
		t.Fatal(err)
	}
	for _, pk := range direct {
		t.Logf("direct: %s via %s: %d bytes %q", pk.From, pk.Interface, len(pk.Data), trim(pk.Data))
	}

	sock := filepath.Join(t.TempDir(), "probe.sock")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	srv := &Server{Socket: sock, Filter: FilterFromEnv(), Log: t.Logf}
	go srv.ListenAndServe(ctx)
	for i := 0; i < 50; i++ {
		if _, err := os.Stat(sock); err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Setenv(SocketEnv, sock)
	relayed, err := Run(context.Background(), p)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("direct %d answers, through the relay %d", len(direct), len(relayed))
	if _, err := Run(context.Background(), Probe{Dest: "192.168.0.14", DestPort: 37810, Payload: []byte("x")}); err == nil {
		t.Fatal("a unicast probe must be refused")
	}
}

func trim(b []byte) string {
	s := string(b)
	if i := strings.Index(s, "{"); i >= 0 {
		s = s[i:]
	}
	if len(s) > 120 {
		s = s[:120]
	}
	return s
}
