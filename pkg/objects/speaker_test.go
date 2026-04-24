package objects

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpeakerObject_Constants(t *testing.T) {
	assert.Equal(t, "speaker.state.idle", SPEAKER_STATE_IDLE)
	assert.Equal(t, "speaker.state.talkback", SPEAKER_STATE_TALKBACK)
	assert.Equal(t, "speaker.action.start_talkback", SPEAKER_ACTION_START_TALKBACK)
	assert.Equal(t, "speaker.action.stop_talkback", SPEAKER_ACTION_STOP_TALKBACK)
}

func TestSpeakerObject_Metadata(t *testing.T) {
	spk := NewSpeakerObject(NewSpeakerObjectProps{
		Metadata: ObjectMetadata{
			ObjectID: "test.speaker.1.out0",
			Name:     "Test Speaker",
			Domain:   "test.speaker",
			DeviceID: "1",
		},
		ProfileToken: "Profile_1",
		OutputToken:  "AudioOutput_1",
		DecoderToken: "AudioDecoder_1",
		Codec:        "G711",
		SampleRate:   8000,
		OutputLevel:  50,
	})

	meta := spk.GetMetadata()
	assert.Equal(t, "test.speaker.1.out0", meta.ObjectID)
	assert.Equal(t, "speaker", meta.Type)
	assert.Equal(t, "test.speaker", meta.Domain)
}

func TestSpeakerObject_Setup(t *testing.T) {
	ctrl := newMockMicController("http://localhost:9999")
	spk := NewSpeakerObject(NewSpeakerObjectProps{
		Metadata:     ObjectMetadata{ObjectID: "spk.1", Domain: "test.speaker"},
		ProfileToken: "Prof1",
		OutputToken:  "Out1",
		DecoderToken: "Dec1",
		Codec:        "G711",
		SampleRate:   8000,
		OutputLevel:  75,
	})

	require.NoError(t, spk.Setup(ctrl))
	assert.Equal(t, SPEAKER_STATE_IDLE, ctrl.getState("spk.1"))
	assert.Equal(t, "G711", ctrl.attrs["spk.1"]["codec"])
	assert.Equal(t, "8000", ctrl.attrs["spk.1"]["sample_rate"])
	assert.Equal(t, "75", ctrl.attrs["spk.1"]["output_level"])
	assert.Equal(t, "Out1", ctrl.attrs["spk.1"]["output_token"])
	assert.Equal(t, "Dec1", ctrl.attrs["spk.1"]["decoder_token"])
}

func TestSpeakerObject_AvailableStatesAndActions(t *testing.T) {
	spk := NewSpeakerObject(NewSpeakerObjectProps{
		Metadata: ObjectMetadata{Domain: "d"},
	})

	states := spk.GetAvailableStates()
	assert.Contains(t, states, SPEAKER_STATE_IDLE)
	assert.Contains(t, states, SPEAKER_STATE_TALKBACK)

	actions := spk.GetAvailableActions()
	require.Len(t, actions, 2)
	assert.Equal(t, SPEAKER_ACTION_START_TALKBACK, actions[0].Action)
	assert.Equal(t, SPEAKER_ACTION_STOP_TALKBACK, actions[1].Action)
}

func TestSpeakerObject_StartTalkback_ConnectsWS(t *testing.T) {
	connCh := make(chan *websocket.Conn, 1)
	srv := newWSSServer(t, connCh)
	defer srv.Close()

	ctrl := newMockMicController(srv.URL)
	talkbackStarted := make(chan string, 1)

	spk := NewSpeakerObject(NewSpeakerObjectProps{
		Metadata:     ObjectMetadata{ObjectID: "spk.1", Domain: "test.speaker"},
		ProfileToken: "Prof1",
		OutputToken:  "Out1",
		DecoderToken: "Dec1",
		Codec:        "G711",
		SampleRate:   8000,
		OutputLevel:  50,
		StartTalkbackFn: func(ctx context.Context, sessionID string, stream *TalkbackStream) error {
			talkbackStarted <- sessionID
			<-ctx.Done()
			return nil
		},
	})
	require.NoError(t, spk.Setup(ctrl))

	payload, _ := json.Marshal(map[string]string{"session_id": "sess-talkback"})
	result, err := spk.RunAction("exec1", SPEAKER_ACTION_START_TALKBACK, payload)
	require.NoError(t, err)
	assert.Equal(t, "true", result["ready"])
	assert.Equal(t, "sess-talkback", result["session_id"])

	select {
	case <-connCh:
	case <-time.After(2 * time.Second):
		t.Fatal("driver did not connect to audio WS server")
	}

	select {
	case sid := <-talkbackStarted:
		assert.Equal(t, "sess-talkback", sid)
	case <-time.After(2 * time.Second):
		t.Fatal("StartTalkbackFn not called")
	}

	assert.Eventually(t, func() bool {
		return ctrl.getState("spk.1") == SPEAKER_STATE_TALKBACK
	}, time.Second, 10*time.Millisecond)
}

func TestSpeakerObject_StopTalkback(t *testing.T) {
	connCh := make(chan *websocket.Conn, 1)
	srv := newWSSServer(t, connCh)
	defer srv.Close()

	ctrl := newMockMicController(srv.URL)
	talkbackDone := make(chan struct{})

	spk := NewSpeakerObject(NewSpeakerObjectProps{
		Metadata: ObjectMetadata{ObjectID: "spk.1", Domain: "d"},
		StartTalkbackFn: func(ctx context.Context, sessionID string, stream *TalkbackStream) error {
			<-stream.Done()
			close(talkbackDone)
			return nil
		},
	})
	require.NoError(t, spk.Setup(ctrl))

	payload, _ := json.Marshal(map[string]string{"session_id": "sess-stop"})
	_, err := spk.RunAction("e1", SPEAKER_ACTION_START_TALKBACK, payload)
	require.NoError(t, err)
	<-connCh

	stopPayload, _ := json.Marshal(map[string]string{"session_id": "sess-stop"})
	_, err = spk.RunAction("e2", SPEAKER_ACTION_STOP_TALKBACK, stopPayload)
	require.NoError(t, err)

	select {
	case <-talkbackDone:
	case <-time.After(2 * time.Second):
		t.Fatal("talkback stream not closed after stop")
	}
}

func TestSpeakerObject_StartTalkback_MissingSessionID(t *testing.T) {
	ctrl := newMockMicController("http://localhost:9999")
	spk := NewSpeakerObject(NewSpeakerObjectProps{
		Metadata: ObjectMetadata{ObjectID: "spk.1"},
	})
	require.NoError(t, spk.Setup(ctrl))

	_, err := spk.RunAction("e1", SPEAKER_ACTION_START_TALKBACK, []byte(`{}`))
	assert.Error(t, err)
}

func TestSpeakerObject_UnknownAction(t *testing.T) {
	ctrl := newMockMicController("http://localhost:9999")
	spk := NewSpeakerObject(NewSpeakerObjectProps{
		Metadata: ObjectMetadata{ObjectID: "spk.1"},
	})
	require.NoError(t, spk.Setup(ctrl))

	_, err := spk.RunAction("e1", "invalid.action", []byte(`{}`))
	assert.Error(t, err)
}

func TestTalkbackStream_ReadRTP(t *testing.T) {
	connCh := make(chan *websocket.Conn, 1)
	srv := newWSSServer(t, connCh)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/audio/stream/test"
	clientWS, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer clientWS.Close()

	serverWS := <-connCh

	// Server (DriversHub) sends talkback RTP
	testData := []byte{0x80, 0x00, 0xAB, 0xCD}
	go func() {
		serverWS.WriteMessage(websocket.BinaryMessage, testData)
	}()

	// Driver reads via TalkbackStream
	stream := &TalkbackStream{ws: clientWS, done: make(chan struct{})}
	received, err := stream.ReadRTP()
	require.NoError(t, err)
	assert.Equal(t, testData, received)
}

func TestTalkbackStream_Close_Idempotent(t *testing.T) {
	connCh := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(nil)
	defer srv.Close()
	_ = connCh

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/audio/stream/test"
	// Use a real ws server from TestMicStream
	srv2 := newWSSServer(t, connCh)
	defer srv2.Close()

	wsURL = "ws" + strings.TrimPrefix(srv2.URL, "http") + "/audio/stream/test"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	<-connCh

	stream := &TalkbackStream{ws: ws, done: make(chan struct{})}
	stream.Close()
	stream.Close()

	select {
	case <-stream.Done():
	default:
		t.Fatal("Done() should be closed")
	}
}

func TestSpeakerObject_StateTransitionsAfterStop(t *testing.T) {
	connCh := make(chan *websocket.Conn, 1)
	srv := newWSSServer(t, connCh)
	defer srv.Close()

	ctrl := newMockMicController(srv.URL)
	spk := NewSpeakerObject(NewSpeakerObjectProps{
		Metadata: ObjectMetadata{ObjectID: "spk.1", Domain: "d"},
		StartTalkbackFn: func(ctx context.Context, sessionID string, stream *TalkbackStream) error {
			<-stream.Done()
			return nil
		},
	})
	require.NoError(t, spk.Setup(ctrl))
	assert.Equal(t, SPEAKER_STATE_IDLE, ctrl.getState("spk.1"))

	payload, _ := json.Marshal(map[string]string{"session_id": "s1"})
	spk.RunAction("e1", SPEAKER_ACTION_START_TALKBACK, payload)
	<-connCh

	assert.Eventually(t, func() bool {
		return ctrl.getState("spk.1") == SPEAKER_STATE_TALKBACK
	}, time.Second, 10*time.Millisecond)

	stopPayload, _ := json.Marshal(map[string]string{"session_id": "s1"})
	spk.RunAction("e2", SPEAKER_ACTION_STOP_TALKBACK, stopPayload)

	assert.Eventually(t, func() bool {
		return ctrl.getState("spk.1") == SPEAKER_STATE_IDLE
	}, time.Second, 10*time.Millisecond)
}
