package objects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimalController satisfies ObjectController for tests that don't need real calls.
type minimalController struct{ mockMicController }

// newTestVideoChannelObject creates a minimal VideoChannelObject wired with the given fn.
func newTestVideoChannelObject(fn func(VideoChannelObject, ObjectController, DownloadVideoClipActionPayload) error) VideoChannelObject {
	return NewVideoChannelObject(NewVideoChannelObjectProps{
		Metadata:            ObjectMetadata{ObjectID: "test-obj", Domain: "test.video_channel"},
		DownloadVideoClipFn: fn,
	})
}

func validDownloadPayload(t *testing.T) []byte {
	t.Helper()
	p := DownloadVideoClipActionPayload{
		ObjectID:   "obj-1",
		ChannelIdx: 0,
		StartTime:  "2026-01-01T00:00:00Z",
		EndTime:    "2026-01-01T00:10:00Z",
		JobID:      "job-abc",
		Timeout:    900,
	}
	b, err := json.Marshal(p)
	require.NoError(t, err)
	return b
}

// TestDownloadAction_appearsInGetAvailableActions verifies the constant is declared and
// the action appears in the list returned by GetAvailableActions.
func TestDownloadAction_appearsInGetAvailableActions(t *testing.T) {
	obj := newTestVideoChannelObject(nil)
	actions := obj.GetAvailableActions()
	var found bool
	for _, a := range actions {
		if a.Action == VIDEO_CHANNEL_ACTION_DOWNLOAD_VIDEO_CLIP {
			found = true
			break
		}
	}
	assert.True(t, found, "VIDEO_CHANNEL_ACTION_DOWNLOAD_VIDEO_CLIP should appear in GetAvailableActions")
}

// TestDownloadAction_runActionCallsFn verifies that RunAction dispatches to downloadVideoClipFn
// and returns {"status":"complete","job_id":...} on success.
func TestDownloadAction_runActionCallsFn(t *testing.T) {
	called := false
	fn := func(_ VideoChannelObject, _ ObjectController, p DownloadVideoClipActionPayload) error {
		called = true
		assert.Equal(t, "job-abc", p.JobID)
		assert.Equal(t, 900, p.Timeout)
		return nil
	}
	obj := newTestVideoChannelObject(fn)

	result, err := obj.RunAction("exec-1", VIDEO_CHANNEL_ACTION_DOWNLOAD_VIDEO_CLIP, validDownloadPayload(t))

	require.NoError(t, err)
	assert.True(t, called, "downloadVideoClipFn should have been called")
	assert.Equal(t, "complete", result["status"])
	assert.Equal(t, "job-abc", result["job_id"])
}

// TestDownloadAction_nilFnReturnsError verifies that a nil DownloadVideoClipFn returns an
// error containing "not supported" rather than panicking.
func TestDownloadAction_nilFnReturnsError(t *testing.T) {
	obj := newTestVideoChannelObject(nil)

	_, err := obj.RunAction("exec-2", VIDEO_CHANNEL_ACTION_DOWNLOAD_VIDEO_CLIP, validDownloadPayload(t))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}

// TestDownloadAction_payloadUnmarshalError verifies that malformed JSON is rejected cleanly.
func TestDownloadAction_payloadUnmarshalError(t *testing.T) {
	fn := func(_ VideoChannelObject, _ ObjectController, _ DownloadVideoClipActionPayload) error {
		t.Fatal("fn should not be called on bad JSON")
		return nil
	}
	obj := newTestVideoChannelObject(fn)

	_, err := obj.RunAction("exec-3", VIDEO_CHANNEL_ACTION_DOWNLOAD_VIDEO_CLIP, []byte("{invalid json"))

	require.Error(t, err)
}

// TestDownloadAction_fnErrorPropagates verifies that an error returned by
// downloadVideoClipFn bubbles up through RunAction unchanged.
func TestDownloadAction_fnErrorPropagates(t *testing.T) {
	sentinel := errors.New("device unreachable")
	fn := func(_ VideoChannelObject, _ ObjectController, _ DownloadVideoClipActionPayload) error {
		return sentinel
	}
	obj := newTestVideoChannelObject(fn)

	_, err := obj.RunAction("exec-4", VIDEO_CHANNEL_ACTION_DOWNLOAD_VIDEO_CLIP, validDownloadPayload(t))

	assert.ErrorIs(t, err, sentinel)
}
