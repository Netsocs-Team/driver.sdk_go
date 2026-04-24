package objects

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	SPEAKER_STATE_IDLE     = "speaker.state.idle"
	SPEAKER_STATE_TALKBACK = "speaker.state.talkback"

	SPEAKER_ACTION_START_TALKBACK = "speaker.action.start_talkback"
	SPEAKER_ACTION_STOP_TALKBACK  = "speaker.action.stop_talkback"
)

// TalkbackStream carries browser audio from DriversHub to the device speaker.
// The driver reads RTP bytes via ReadRTP and forwards them to the device backchannel.
type TalkbackStream struct {
	ws   *websocket.Conn
	done chan struct{}
	once sync.Once
}

// ReadRTP blocks until a raw RTP packet is received from DriversHub.
// Returns the raw bytes or an error if the stream is closed.
func (s *TalkbackStream) ReadRTP() ([]byte, error) {
	_, data, err := s.ws.ReadMessage()
	return data, err
}

// Done returns a channel that is closed when the stream ends.
func (s *TalkbackStream) Done() <-chan struct{} { return s.done }

// Close terminates the stream and the underlying WebSocket connection.
func (s *TalkbackStream) Close() {
	s.once.Do(func() {
		s.ws.Close()
		close(s.done)
	})
}

// SpeakerObject represents a physical speaker (backchannel) on a device registered in the Netsocs platform.
type SpeakerObject interface {
	RegistrableObject
	SetStateTalkback() error
	SetStateIdle() error
}

// NewSpeakerObjectProps holds the configuration required to create a SpeakerObject.
type NewSpeakerObjectProps struct {
	Metadata     ObjectMetadata
	ProfileToken string // device media profile token
	OutputToken  string // device audio output token
	DecoderToken string // device audio decoder token (backchannel)
	Codec        string // e.g. "G711"
	SampleRate   int    // e.g. 8000
	OutputLevel  int    // volume 0-100

	// StartTalkbackFn is called when a client starts speaking to this speaker.
	// The SDK opens the WebSocket to DriversHub and passes a ready TalkbackStream.
	// The function must call stream.ReadRTP in a loop, forward each RTP packet to
	// the device backchannel, and return only when ctx is cancelled or stream.Done() is closed.
	StartTalkbackFn func(ctx context.Context, sessionID string, stream *TalkbackStream) error
}

type speakerObject struct {
	props          NewSpeakerObjectProps
	controller     ObjectController
	activeSessions sync.Map // sessionID → *TalkbackStream
}

func (s *speakerObject) GetMetadata() ObjectMetadata {
	s.props.Metadata.Type = "speaker"
	return s.props.Metadata
}

func (s *speakerObject) SetState(st string) error {
	return s.controller.SetState(s.props.Metadata.ObjectID, st)
}

func (s *speakerObject) UpdateStateAttributes(a map[string]string) error {
	return s.controller.UpdateStateAttributes(s.props.Metadata.ObjectID, a)
}

func (s *speakerObject) AddEventTypes(eventTypes []EventType) error {
	return s.controller.AddEventTypes(eventTypes)
}

func (s *speakerObject) SetStateTalkback() error {
	return s.controller.SetState(s.props.Metadata.ObjectID, SPEAKER_STATE_TALKBACK)
}

func (s *speakerObject) SetStateIdle() error {
	return s.controller.SetState(s.props.Metadata.ObjectID, SPEAKER_STATE_IDLE)
}

func (s *speakerObject) GetAvailableStates() []string {
	return []string{SPEAKER_STATE_IDLE, SPEAKER_STATE_TALKBACK}
}

func (s *speakerObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{
		{Action: SPEAKER_ACTION_START_TALKBACK, Domain: s.props.Metadata.Domain},
		{Action: SPEAKER_ACTION_STOP_TALKBACK, Domain: s.props.Metadata.Domain},
	}
}

func (s *speakerObject) Setup(ctrl ObjectController) error {
	s.controller = ctrl
	_ = ctrl.UpdateStateAttributes(s.props.Metadata.ObjectID, map[string]string{
		"profile_token": s.props.ProfileToken,
		"output_token":  s.props.OutputToken,
		"decoder_token": s.props.DecoderToken,
		"codec":         s.props.Codec,
		"sample_rate":   fmt.Sprint(s.props.SampleRate),
		"output_level":  fmt.Sprint(s.props.OutputLevel),
	})
	return ctrl.SetState(s.props.Metadata.ObjectID, SPEAKER_STATE_IDLE)
}

func (s *speakerObject) RunAction(executionID, action string, payload []byte) (map[string]string, error) {
	switch action {
	case SPEAKER_ACTION_START_TALKBACK:
		return s.startTalkback(payload)
	case SPEAKER_ACTION_STOP_TALKBACK:
		return s.stopTalkback(payload)
	}
	return nil, fmt.Errorf("speaker: unknown action %q", action)
}

func (s *speakerObject) startTalkback(payload []byte) (map[string]string, error) {
	var p struct {
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(payload, &p); err != nil || p.SessionID == "" {
		return nil, fmt.Errorf("speaker: start_talkback requires session_id")
	}

	wsURL := buildAudioStreamURL(s.controller.GetDriverhubHost(), p.SessionID)
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, wsDriverAuthHeader(s.controller.GetDriverKey()))
	if err != nil {
		return nil, fmt.Errorf("speaker: dial DriversHub audio stream: %w", err)
	}

	stream := &TalkbackStream{ws: ws, done: make(chan struct{})}
	s.activeSessions.Store(p.SessionID, stream)
	_ = s.SetStateTalkback()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		defer func() {
			s.activeSessions.Delete(p.SessionID)
			stream.Close()
			_ = s.SetStateIdle()
		}()
		if s.props.StartTalkbackFn != nil {
			s.props.StartTalkbackFn(ctx, p.SessionID, stream)
		}
	}()

	return map[string]string{"ready": "true", "session_id": p.SessionID}, nil
}

func (s *speakerObject) stopTalkback(payload []byte) (map[string]string, error) {
	var p struct {
		SessionID string `json:"session_id"`
	}
	_ = json.Unmarshal(payload, &p)
	if v, ok := s.activeSessions.LoadAndDelete(p.SessionID); ok {
		v.(*TalkbackStream).Close()
	}
	return nil, nil
}

// NewSpeakerObject creates a SpeakerObject ready to be registered with the SDK client.
func NewSpeakerObject(props NewSpeakerObjectProps) SpeakerObject {
	return &speakerObject{props: props}
}
