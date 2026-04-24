package objects

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- mock controller ----

type mockMicController struct {
	mu        sync.Mutex
	hubHost   string
	driverKey string
	states    map[string]string
	attrs     map[string]map[string]string
}

func newMockMicController(serverURL string) *mockMicController {
	return &mockMicController{
		hubHost:   serverURL,
		driverKey: "test-driver-key",
		states:    make(map[string]string),
		attrs:     make(map[string]map[string]string),
	}
}

func (m *mockMicController) GetDriverhubHost() string { return m.hubHost }
func (m *mockMicController) GetDriverKey() string     { return m.driverKey }
func (m *mockMicController) SetState(id, state string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[id] = state
	return nil
}
func (m *mockMicController) GetState(id string) (StateRecord, error) { return StateRecord{}, nil }
func (m *mockMicController) UpdateStateAttributes(id string, a map[string]string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attrs[id] = a
	return nil
}
func (m *mockMicController) UpdateResultAttributes(execID string, a map[string]string) error {
	return nil
}
func (m *mockMicController) NewAction(action ObjectAction) error           { return nil }
func (m *mockMicController) CreateObject(obj RegistrableObject) error      { return nil }
func (m *mockMicController) ListenActionRequests() error                   { return nil }
func (m *mockMicController) DisabledObject(id string) error                { return nil }
func (m *mockMicController) EnabledObject(id string) error                 { return nil }
func (m *mockMicController) AddEventTypes(et []EventType) error            { return nil }
func (m *mockMicController) Increment(id string) error                     { return nil }
func (m *mockMicController) Decrement(id string) error                     { return nil }

func (m *mockMicController) getState(id string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.states[id]
}

// ---- WS test server helper ----

var wsUpgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

// newWSSServer starts an httptest server that accepts WS on /audio/stream/:id.
// connCh receives every accepted *websocket.Conn.
func newWSSServer(t *testing.T, connCh chan *websocket.Conn) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/audio/stream/") {
			http.NotFound(w, r)
			return
		}
		ws, err := wsUpgrader.Upgrade(w, r, nil)
		require.NoError(t, err)
		connCh <- ws
	}))
	return srv
}

// ---- tests ----

func TestMicrophoneObject_Constants(t *testing.T) {
	assert.Equal(t, "microphone.state.idle", MICROPHONE_STATE_IDLE)
	assert.Equal(t, "microphone.state.streaming", MICROPHONE_STATE_STREAMING)
	assert.Equal(t, "microphone.action.start_stream", MICROPHONE_ACTION_START_STREAM)
	assert.Equal(t, "microphone.action.stop_stream", MICROPHONE_ACTION_STOP_STREAM)
}

func TestMicrophoneObject_Metadata(t *testing.T) {
	mic := NewMicrophoneObject(NewMicrophoneObjectProps{
		Metadata: ObjectMetadata{
			ObjectID: "test.microphone.1.src0",
			Name:     "Test Mic",
			Domain:   "test.microphone",
			DeviceID: "1",
		},
		ProfileToken: "Profile_1",
		SourceToken:  "AudioSource_1",
		Codec:        "G711",
		SampleRate:   8000,
		Channels:     1,
	})

	meta := mic.GetMetadata()
	assert.Equal(t, "test.microphone.1.src0", meta.ObjectID)
	assert.Equal(t, "microphone", meta.Type)
	assert.Equal(t, "test.microphone", meta.Domain)
}

func TestMicrophoneObject_Setup(t *testing.T) {
	ctrl := newMockMicController("http://localhost:9999")
	mic := NewMicrophoneObject(NewMicrophoneObjectProps{
		Metadata:     ObjectMetadata{ObjectID: "mic.1", Domain: "test.microphone"},
		ProfileToken: "Prof1",
		SourceToken:  "Src1",
		Codec:        "G711",
		SampleRate:   8000,
		Channels:     1,
	})

	require.NoError(t, mic.Setup(ctrl))
	assert.Equal(t, MICROPHONE_STATE_IDLE, ctrl.getState("mic.1"))
	assert.Equal(t, "G711", ctrl.attrs["mic.1"]["codec"])
	assert.Equal(t, "8000", ctrl.attrs["mic.1"]["sample_rate"])
}

func TestMicrophoneObject_AvailableStatesAndActions(t *testing.T) {
	mic := NewMicrophoneObject(NewMicrophoneObjectProps{
		Metadata: ObjectMetadata{Domain: "d"},
	})

	states := mic.GetAvailableStates()
	assert.Contains(t, states, MICROPHONE_STATE_IDLE)
	assert.Contains(t, states, MICROPHONE_STATE_STREAMING)

	actions := mic.GetAvailableActions()
	require.Len(t, actions, 2)
	assert.Equal(t, MICROPHONE_ACTION_START_STREAM, actions[0].Action)
	assert.Equal(t, MICROPHONE_ACTION_STOP_STREAM, actions[1].Action)
}

func TestMicrophoneObject_StartStream_ConnectsWS(t *testing.T) {
	connCh := make(chan *websocket.Conn, 1)
	srv := newWSSServer(t, connCh)
	defer srv.Close()

	ctrl := newMockMicController(srv.URL)
	streamStarted := make(chan string, 1)

	mic := NewMicrophoneObject(NewMicrophoneObjectProps{
		Metadata:     ObjectMetadata{ObjectID: "mic.1", Domain: "test.microphone"},
		ProfileToken: "Prof1",
		SourceToken:  "Src1",
		Codec:        "G711",
		SampleRate:   8000,
		Channels:     1,
		StartStreamFn: func(ctx context.Context, sessionID string, stream *MicStream) error {
			streamStarted <- sessionID
			<-ctx.Done()
			return nil
		},
	})
	require.NoError(t, mic.Setup(ctrl))

	payload, _ := json.Marshal(map[string]string{"session_id": "sess-abc"})
	result, err := mic.RunAction("exec1", MICROPHONE_ACTION_START_STREAM, payload)
	require.NoError(t, err)
	assert.Equal(t, "true", result["ready"])
	assert.Equal(t, "sess-abc", result["session_id"])

	// Server receives WS connection
	select {
	case <-connCh:
	case <-time.After(2 * time.Second):
		t.Fatal("driver did not connect to audio WS server")
	}

	// StartStreamFn is called
	select {
	case sid := <-streamStarted:
		assert.Equal(t, "sess-abc", sid)
	case <-time.After(2 * time.Second):
		t.Fatal("StartStreamFn not called")
	}

	// State transitions to streaming
	assert.Eventually(t, func() bool {
		return ctrl.getState("mic.1") == MICROPHONE_STATE_STREAMING
	}, time.Second, 10*time.Millisecond)
}

func TestMicrophoneObject_StopStream(t *testing.T) {
	connCh := make(chan *websocket.Conn, 1)
	srv := newWSSServer(t, connCh)
	defer srv.Close()

	ctrl := newMockMicController(srv.URL)
	streamDone := make(chan struct{})

	mic := NewMicrophoneObject(NewMicrophoneObjectProps{
		Metadata:    ObjectMetadata{ObjectID: "mic.1", Domain: "d"},
		StartStreamFn: func(ctx context.Context, sessionID string, stream *MicStream) error {
			<-stream.Done()
			close(streamDone)
			return nil
		},
	})
	require.NoError(t, mic.Setup(ctrl))

	payload, _ := json.Marshal(map[string]string{"session_id": "sess-xyz"})
	_, err := mic.RunAction("e1", MICROPHONE_ACTION_START_STREAM, payload)
	require.NoError(t, err)
	<-connCh // wait for WS connect

	stopPayload, _ := json.Marshal(map[string]string{"session_id": "sess-xyz"})
	_, err = mic.RunAction("e2", MICROPHONE_ACTION_STOP_STREAM, stopPayload)
	require.NoError(t, err)

	select {
	case <-streamDone:
	case <-time.After(2 * time.Second):
		t.Fatal("stream not closed after stop")
	}
}

func TestMicrophoneObject_StartStream_MissingSessionID(t *testing.T) {
	ctrl := newMockMicController("http://localhost:9999")
	mic := NewMicrophoneObject(NewMicrophoneObjectProps{
		Metadata: ObjectMetadata{ObjectID: "mic.1"},
	})
	require.NoError(t, mic.Setup(ctrl))

	_, err := mic.RunAction("e1", MICROPHONE_ACTION_START_STREAM, []byte(`{}`))
	assert.Error(t, err)
}

func TestMicrophoneObject_UnknownAction(t *testing.T) {
	ctrl := newMockMicController("http://localhost:9999")
	mic := NewMicrophoneObject(NewMicrophoneObjectProps{
		Metadata: ObjectMetadata{ObjectID: "mic.1"},
	})
	require.NoError(t, mic.Setup(ctrl))

	_, err := mic.RunAction("e1", "unknown.action", []byte(`{}`))
	assert.Error(t, err)
}

func TestMicStream_WriteRTP(t *testing.T) {
	connCh := make(chan *websocket.Conn, 1)
	srv := newWSSServer(t, connCh)
	defer srv.Close()

	// Connect a WS client
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/audio/stream/test"
	clientWS, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer clientWS.Close()

	serverWS := <-connCh

	// Write from client side (simulating driver writing RTP)
	stream := &MicStream{ws: clientWS, done: make(chan struct{})}
	testData := []byte{0x80, 0x00, 0x01, 0x02, 0x03}
	require.NoError(t, stream.WriteRTP(testData))

	// Server (DriversHub) receives it
	_, received, err := serverWS.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, testData, received)
}

func TestMicStream_Close_Idempotent(t *testing.T) {
	connCh := make(chan *websocket.Conn, 1)
	srv := newWSSServer(t, connCh)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/audio/stream/test"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	<-connCh

	stream := &MicStream{ws: ws, done: make(chan struct{})}
	stream.Close()
	stream.Close() // second call must not panic

	select {
	case <-stream.Done():
	default:
		t.Fatal("Done() should be closed")
	}
}

func TestBuildAudioStreamURL(t *testing.T) {
	cases := []struct {
		hub      string
		session  string
		expected string
	}{
		{"http://10.0.0.1:3196", "abc", "ws://10.0.0.1:3196/audio/stream/abc"},
		{"https://hub.example.com", "xyz", "wss://hub.example.com/audio/stream/xyz"},
		{"http://hub:8080/api/netsocs/dh", "s1", "ws://hub:8080/audio/stream/s1"},
	}
	for _, c := range cases {
		got := buildAudioStreamURL(c.hub, c.session)
		assert.Equal(t, c.expected, got, "hub=%q session=%q", c.hub, c.session)
	}
}
