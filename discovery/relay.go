package discovery

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

// The protocol on the socket: one JSON request line, one JSON response line,
// then the connection closes.
type response struct {
	Packets []Packet `json:"packets"`
	Error   string   `json:"error,omitempty"`
}

const maxRequest = 16 << 10

func runRemote(ctx context.Context, sock string, p Probe) ([]Packet, error) {
	d := net.Dialer{Timeout: 2 * time.Second}
	c, err := d.DialContext(ctx, "unix", sock)
	if err != nil {
		return nil, fmt.Errorf("discovery: the netsocs-discovery service does not answer at %s: %w", sock, err)
	}
	defer c.Close()
	// The service may wait for another scan on the same port first.
	_ = c.SetDeadline(time.Now().Add(2*p.Timeout + 5*time.Second))
	req, _ := json.Marshal(p)
	if _, err := c.Write(append(req, '\n')); err != nil {
		return nil, fmt.Errorf("discovery: send to the service: %w", err)
	}
	line, err := bufio.NewReaderSize(c, 1<<20).ReadBytes('\n')
	if err != nil && !(errors.Is(err, io.EOF) && len(line) > 0) {
		return nil, fmt.Errorf("discovery: read the service's answer: %w", err)
	}
	var res response
	if err := json.Unmarshal(line, &res); err != nil {
		return nil, fmt.Errorf("discovery: bad answer from the service: %w", err)
	}
	if res.Error != "" {
		return nil, errors.New(res.Error)
	}
	return res.Packets, nil
}

// Server is the netsocs-discovery service: it runs the probes the drivers
// send on the socket, on the host network.
type Server struct {
	Socket string
	Filter InterfaceFilter
	Log    func(format string, args ...any)
	// MaxConcurrent caps the scans running at once (default 4).
	MaxConcurrent int
}

// ListenAndServe creates the socket (any local process that can reach the
// file may scan: the directory is shared only with the site container) and
// serves until ctx ends.
func (s *Server) ListenAndServe(ctx context.Context) error {
	if s.Log == nil {
		s.Log = func(string, ...any) {}
	}
	if s.MaxConcurrent <= 0 {
		s.MaxConcurrent = 4
	}
	if err := os.MkdirAll(filepath.Dir(s.Socket), 0o755); err != nil {
		return err
	}
	_ = os.Remove(s.Socket)
	l, err := net.Listen("unix", s.Socket)
	if err != nil {
		return err
	}
	defer os.Remove(s.Socket)
	if err := os.Chmod(s.Socket, 0o666); err != nil {
		l.Close()
		return err
	}
	go func() { <-ctx.Done(); l.Close() }()
	slots := make(chan struct{}, s.MaxConcurrent)
	for {
		c, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		go func() {
			slots <- struct{}{}
			defer func() { <-slots }()
			s.handle(ctx, c)
		}()
	}
}

func (s *Server) handle(ctx context.Context, c net.Conn) {
	defer c.Close()
	_ = c.SetReadDeadline(time.Now().Add(5 * time.Second))
	line, err := bufio.NewReaderSize(io.LimitReader(c, maxRequest), maxRequest).ReadBytes('\n')
	if err != nil && !(errors.Is(err, io.EOF) && len(line) > 0) {
		return
	}
	var p Probe
	res := response{}
	if err := json.Unmarshal(line, &p); err != nil {
		res.Error = "discovery: bad request"
	} else if err := p.Validate(); err != nil {
		res.Error = err.Error()
		s.Log("refused probe to %s:%d: %v", p.Dest, p.DestPort, err)
	} else {
		pkts, err := Execute(ctx, p, s.Filter)
		if err != nil {
			res.Error = err.Error()
			s.Log("probe to %s:%d failed: %v", p.Dest, p.DestPort, err)
		} else {
			res.Packets = pkts
			s.Log("probe to %s:%d (answers on %d): %d answers", p.Dest, p.DestPort, p.ListenPort, len(pkts))
		}
	}
	b, _ := json.Marshal(res)
	_ = c.SetWriteDeadline(time.Now().Add(5 * time.Second))
	_, _ = c.Write(append(b, '\n'))
}
