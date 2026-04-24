package objects

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	MICROPHONE_STATE_IDLE      = "microphone.state.idle"
	MICROPHONE_STATE_STREAMING = "microphone.state.streaming"

	MICROPHONE_ACTION_START_STREAM = "microphone.action.start_stream"
	MICROPHONE_ACTION_STOP_STREAM  = "microphone.action.stop_stream"
)

// MicStream is an active audio stream from a device microphone to DriversHub.
// The driver reads RTP bytes from the device and calls WriteRTP to forward them.
// It blocks until ctx is cancelled or the stream is closed remotely.
type MicStream struct {
	ws   *websocket.Conn
	done chan struct{}
	once sync.Once
}

// WriteRTP forwards raw RTP bytes to DriversHub over the WebSocket connection.
func (s *MicStream) WriteRTP(data []byte) error {
	return s.ws.WriteMessage(websocket.BinaryMessage, data)
}

// Done returns a channel that is closed when the stream ends.
func (s *MicStream) Done() <-chan struct{} { return s.done }

// Close terminates the stream and the underlying WebSocket connection.
func (s *MicStream) Close() {
	s.once.Do(func() {
		s.ws.Close()
		close(s.done)
	})
}

// MicrophoneObject represents a physical microphone on a device registered in the Netsocs platform.
type MicrophoneObject interface {
	RegistrableObject
	SetStateStreaming() error
	SetStateIdle() error
}

// NewMicrophoneObjectProps holds the configuration required to create a MicrophoneObject.
type NewMicrophoneObjectProps struct {
	Metadata     ObjectMetadata
	ProfileToken string // device media profile token
	SourceToken  string // device audio source token
	Codec        string // e.g. "G711", "G726", "AAC"
	SampleRate   int    // e.g. 8000
	Channels     int    // e.g. 1

	// StartStreamFn is called when a client requests audio from this microphone.
	// The SDK opens the WebSocket to DriversHub and passes a ready MicStream.
	// The function must read RTP audio from the device, call stream.WriteRTP in a loop,
	// and return only when ctx is cancelled or stream.Done() is closed.
	StartStreamFn func(ctx context.Context, sessionID string, stream *MicStream) error
}

type microphoneObject struct {
	props         NewMicrophoneObjectProps
	controller    ObjectController
	activeStreams sync.Map // sessionID → *MicStream
}

func (m *microphoneObject) GetMetadata() ObjectMetadata {
	m.props.Metadata.Type = "microphone"
	return m.props.Metadata
}

func (m *microphoneObject) SetState(s string) error {
	return m.controller.SetState(m.props.Metadata.ObjectID, s)
}

func (m *microphoneObject) UpdateStateAttributes(a map[string]string) error {
	return m.controller.UpdateStateAttributes(m.props.Metadata.ObjectID, a)
}

func (m *microphoneObject) AddEventTypes(eventTypes []EventType) error {
	return m.controller.AddEventTypes(eventTypes)
}

func (m *microphoneObject) SetStateStreaming() error {
	return m.controller.SetState(m.props.Metadata.ObjectID, MICROPHONE_STATE_STREAMING)
}

func (m *microphoneObject) SetStateIdle() error {
	return m.controller.SetState(m.props.Metadata.ObjectID, MICROPHONE_STATE_IDLE)
}

func (m *microphoneObject) GetAvailableStates() []string {
	return []string{MICROPHONE_STATE_IDLE, MICROPHONE_STATE_STREAMING}
}

func (m *microphoneObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{
		{Action: MICROPHONE_ACTION_START_STREAM, Domain: m.props.Metadata.Domain},
		{Action: MICROPHONE_ACTION_STOP_STREAM, Domain: m.props.Metadata.Domain},
	}
}

func (m *microphoneObject) Setup(ctrl ObjectController) error {
	m.controller = ctrl
	_ = ctrl.UpdateStateAttributes(m.props.Metadata.ObjectID, map[string]string{
		"profile_token": m.props.ProfileToken,
		"source_token":  m.props.SourceToken,
		"codec":         m.props.Codec,
		"sample_rate":   fmt.Sprint(m.props.SampleRate),
		"channels":      fmt.Sprint(m.props.Channels),
	})
	return ctrl.SetState(m.props.Metadata.ObjectID, MICROPHONE_STATE_IDLE)
}

func (m *microphoneObject) RunAction(executionID, action string, payload []byte) (map[string]string, error) {
	switch action {
	case MICROPHONE_ACTION_START_STREAM:
		return m.startStream(payload)
	case MICROPHONE_ACTION_STOP_STREAM:
		return m.stopStream(payload)
	}
	return nil, fmt.Errorf("microphone: unknown action %q", action)
}

func (m *microphoneObject) startStream(payload []byte) (map[string]string, error) {
	var p struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil || p.SessionID == "" {
		return nil, fmt.Errorf("microphone: start_stream requires session_id")
	}

	wsURL := buildAudioStreamURL(m.controller.GetDriverhubHost(), p.SessionID)
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, wsDriverAuthHeader(m.controller.GetDriverKey()))
	if err != nil {
		return nil, fmt.Errorf("microphone: dial DriversHub audio stream: %w", err)
	}

	stream := &MicStream{ws: ws, done: make(chan struct{})}
	m.activeStreams.Store(p.SessionID, stream)
	_ = m.SetStateStreaming()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		defer func() {
			m.activeStreams.Delete(p.SessionID)
			stream.Close()
			_ = m.SetStateIdle()
		}()
		if m.props.StartStreamFn != nil {
			m.props.StartStreamFn(ctx, p.SessionID, stream)
		}
	}()

	return map[string]string{"ready": "true", "session_id": p.SessionID}, nil
}

func (m *microphoneObject) stopStream(payload []byte) (map[string]string, error) {
	var p struct {
		SessionID string `json:"session_id"`
	}
	_ = json.Unmarshal(payload, &p)
	if v, ok := m.activeStreams.LoadAndDelete(p.SessionID); ok {
		v.(*MicStream).Close()
	}
	return nil, nil
}

// NewMicrophoneObject creates a MicrophoneObject ready to be registered with the SDK client.
func NewMicrophoneObject(props NewMicrophoneObjectProps) MicrophoneObject {
	return &microphoneObject{props: props}
}
