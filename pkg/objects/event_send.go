package objects

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

// EventSendTimeout bounds one attempt to send an event to the central, from
// dial to the last byte of the answer.
const EventSendTimeout = 10 * time.Second

// EventNotStoredError: the central did not confirm the event was stored. The
// event may be retried with the same ID; the central does not store it twice.
type EventNotStoredError struct {
	ID         string
	StatusCode int // 0 when no answer came back (network error, timeout)
	Body       string
	Err        error
}

func (e *EventNotStoredError) Error() string {
	if e.StatusCode == 0 {
		return fmt.Sprintf("event %s not stored: %v", e.ID, e.Err)
	}
	return fmt.Sprintf("event %s not stored: central answered %d: %s", e.ID, e.StatusCode, e.Body)
}

func (e *EventNotStoredError) Unwrap() error { return e.Err }

// Retryable reports whether sending the same event again may succeed: no
// answer, 429 or a 5xx. A 4xx means the event itself was refused.
func (e *EventNotStoredError) Retryable() bool {
	return e.StatusCode == 0 || e.StatusCode == 429 || e.StatusCode >= 500
}

// NewEventID returns a new event id: a UUID version 7 (RFC 9562), so ids
// sort by creation time.
func NewEventID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	ms := uint64(time.Now().UnixMilli())
	var ts [8]byte
	binary.BigEndian.PutUint64(ts[:], ms)
	copy(b[0:6], ts[2:8])
	b[6] = (b[6] & 0x0f) | 0x70 // version 7
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 9562
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
