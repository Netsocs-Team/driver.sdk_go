package objects

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// emptyVideoChannel returns an object with no action functions set (all nil).
func emptyVideoChannel() VideoChannelObject {
	return NewVideoChannelObject(NewVideoChannelObjectProps{
		Metadata: ObjectMetadata{ObjectID: "nil-test", Domain: "test.video_channel"},
	})
}

func mustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

// TestRunAction_nilFns verifies that every action case returns an error containing
// "not supported" (instead of panicking) when the corresponding function is nil.
func TestRunAction_nilFns(t *testing.T) {
	obj := emptyVideoChannel()

	cases := []struct {
		action  string
		payload []byte
	}{
		{
			action:  VIDEO_CHANNEL_ACTION_SNAPSHOT,
			payload: mustMarshal(t, SnapshotActionPayload{}),
		},
		{
			action:  VIDEO_CHANNEL_ACTION_VIDEOCLIP,
			payload: mustMarshal(t, VideoClipActionPayload{}),
		},
		{
			action:  VIDEO_CHANNEL_ACTION_PTZ_CONTROL,
			payload: mustMarshal(t, VideoChannelActionPtzControlPayload{}),
		},
		{
			action:  VIDEO_CHANNEL_ACTION_PTZ_GOTO_PRESET,
			payload: mustMarshal(t, VideoChannelActionPtzGotoPresetPayload{}),
		},
		{
			action:  VIDEO_CHANNEL_ACTION_SEEK,
			payload: mustMarshal(t, SeekPayload{}),
		},
		{
			action:  VIDEO_CHANNEL_ACTION_REQUEST_DOLYNK_STREAM_URL,
			payload: mustMarshal(t, RequestDolynkStreamURLPayload{}),
		},
		{
			action:  VIDEO_CHANNEL_ACTION_REQUEST_DAHUA_PLAYBACK_MEDIA_FILES,
			payload: mustMarshal(t, RequestDahuaPlaybackMediaFilesPayload{}),
		},
		{
			action:  VIDEO_CHANNEL_ACTION_GET_RECORDING_SEGMENTS,
			payload: mustMarshal(t, GetRecordingSegmentsPayload{}),
		},
		{
			action:  VIDEO_CHANNEL_ACTION_PTZ_GET_STATUS,
			payload: []byte(`{}`),
		},
		{
			action:  VIDEO_CHANNEL_ACTION_PUBLISH_STREAM_START,
			payload: mustMarshal(t, PublishStreamStartPayload{}),
		},
		{
			action:  VIDEO_CHANNEL_ACTION_PUBLISH_STREAM_STOP,
			payload: mustMarshal(t, PublishStreamStopPayload{}),
		},
		{
			action:  VIDEO_CHANNEL_ACTION_DOWNLOAD_VIDEO_CLIP,
			payload: mustMarshal(t, DownloadVideoClipActionPayload{}),
		},
	}

	for _, tc := range cases {
		t.Run(tc.action, func(t *testing.T) {
			_, err := obj.RunAction("exec", tc.action, tc.payload)
			require.Error(t, err, "expected error, got nil for action %s", tc.action)
			assert.Contains(t, err.Error(), "not supported", "action %s: error should mention 'not supported'", tc.action)
		})
	}
}
