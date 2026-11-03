package objects

import (
	"encoding/json"
	"testing"
)

// DriverHub adds publish_token to publish_stream_start and to playback actions
// (DriversHub whip/publish_auth.go). The driver must see it to send it as the
// WHIP Bearer token.
func TestPublishTokenReachesThePayloads(t *testing.T) {
	var start PublishStreamStartPayload
	if err := json.Unmarshal([]byte(`{"quality":"sub","publish_token":"tok-1"}`), &start); err != nil {
		t.Fatal(err)
	}
	if start.PublishToken != "tok-1" || start.Quality != "sub" {
		t.Errorf("publish_stream_start: got %+v", start)
	}

	var seek SeekPayload
	if err := json.Unmarshal([]byte(`{"playback_id":"pb-1","publish_token":"tok-2"}`), &seek); err != nil {
		t.Fatal(err)
	}
	if seek.PublishToken != "tok-2" || seek.PlaybackID != "pb-1" {
		t.Errorf("seek: got %+v", seek)
	}

	// Older DriverHub: no token, field empty, nothing else changes.
	var old PublishStreamStartPayload
	if err := json.Unmarshal([]byte(`{"quality":"main"}`), &old); err != nil || old.PublishToken != "" {
		t.Errorf("payload without token: %+v, %v", old, err)
	}
}
