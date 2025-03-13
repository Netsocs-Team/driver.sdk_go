package objects

import (
	"strconv"

	"github.com/goccy/go-json"
)

const VIDEO_CHANNEL_STATE_STREAMING = "video_channel.state.streaming"
const VIDEO_CHANNEL_STATE_RECORDING = "video_channel.state.recording"
const VIDEO_CHANNEL_STATE_IDLE = "video_channel.state.idle"
const VIDEO_CHANNEL_STATE_UNKNOWN = "video_channel.state.unknown"

const VIDEO_CHANNEL_ACTION_SNAPSHOT = "video_channel.action.snapshot"
const VIDEO_CHANNEL_ACTION_VIDEOCLIP = "video_channel.action.videoclip"
const VIDEO_CHANNEL_ACTION_PTZ_CONTROL = "video_channel.action.ptz_control"

type VideoChannelActionPtzControlPayload struct {
	Pan      int  `json:"pan"`
	Tilt     int  `json:"tilt"`
	Zoom     int  `json:"zoom"`
	Relative bool `json:"relative"`
}

type VideoClipActionPayload struct {
	StartTimestamp string `json:"start_timestamp"`
	EndTimestamp   string `json:"end_timestamp"`
	Resolution     string `json:"resolution,omitempty"` //"1920x1080"
	Timeout        int    `json:"timeout,omitempty"`    // This is to stop trying to make the video clip after certain minutes
}

type SnapshotActionPayload struct {
	Timestamp  string `json:"timestamp.omitempty"`  //if its empty make it as soon as received
	Resolution string `json:"resolution,omitempty"` //"1920x1080"
}

type VideoChannelActionPtzControlPayloadDirection string

type VideoChannelObject interface {
	RegistrableObject
	// stream helpers
	SecondaryStream(streamId string) error
	PrimaryStream(streamId string) error
	// state helpers
	SetModeRecording() error
	SetModeIdle() error
	SetModeStreaming() error
	SetModeUnknown() error
}

type videoChannelObject struct {
	setupFn    func(VideoChannelObject, ObjectController) error
	controller ObjectController
	metadata   ObjectMetadata

	streamId      string
	subStreamId   string
	ptz           bool
	videoEngineId string
	// actions functions
	snapshotFn  func(VideoChannelObject, ObjectController, SnapshotActionPayload) (filename string, err error)
	videoclipFn func(VideoChannelObject, ObjectController, VideoClipActionPayload) (filename string, err error)
	ptzFn       func(VideoChannelObject, ObjectController, VideoChannelActionPtzControlPayload) error
}

// UpdateStateAttributes implements VideoChannelObject.
func (v *videoChannelObject) UpdateStateAttributes(attributes map[string]string) error {
	return v.controller.UpdateStateAttributes(v.GetMetadata().ObjectID, attributes)
}

// SetState implements VideoChannelObject.
func (v *videoChannelObject) SetState(state string) error {
	panic("unimplemented")
}

// AddEventTypes implements VideoChannelObject.
func (v *videoChannelObject) AddEventTypes(eventTypes []EventType) error {
	panic("unimplemented")
}

// SetModeIdle implements VideoChannelObject.
func (v *videoChannelObject) SetModeIdle() error {
	return v.controller.SetState(v.GetMetadata().ObjectID, VIDEO_CHANNEL_STATE_IDLE)
}

// SetModeRecording implements VideoChannelObject.
func (v *videoChannelObject) SetModeRecording() error {
	return v.controller.SetState(v.GetMetadata().ObjectID, VIDEO_CHANNEL_STATE_RECORDING)
}

// SetModeStreaming implements VideoChannelObject.
func (v *videoChannelObject) SetModeStreaming() error {
	return v.controller.SetState(v.GetMetadata().ObjectID, VIDEO_CHANNEL_STATE_STREAMING)
}

// SetModeUnknown implements VideoChannelObject.
func (v *videoChannelObject) SetModeUnknown() error {
	return v.controller.SetState(v.GetMetadata().ObjectID, VIDEO_CHANNEL_STATE_UNKNOWN)
}

// PrimaryStream implements VideoChannelObject.
func (v *videoChannelObject) PrimaryStream(streamId string) error {
	return v.controller.UpdateStateAttributes(v.GetMetadata().ObjectID, map[string]string{"primary_stream": streamId})
}

// SecundaryStream implements VideoChannelObject.
func (v *videoChannelObject) SecondaryStream(streamId string) error {
	return v.controller.UpdateStateAttributes(v.GetMetadata().ObjectID, map[string]string{"secondary_stream": streamId})
}

// GetAvailableActions implements VideoChannelObject.
func (v *videoChannelObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{
		{Action: VIDEO_CHANNEL_ACTION_SNAPSHOT, Domain: v.metadata.Domain},
		{Action: VIDEO_CHANNEL_ACTION_PTZ_CONTROL, Domain: v.metadata.Domain},
		{Action: VIDEO_CHANNEL_ACTION_VIDEOCLIP, Domain: v.metadata.Domain},
	}
}

// GetAvailableStates implements VideoChannelObject.
func (v *videoChannelObject) GetAvailableStates() []string {
	return []string{
		VIDEO_CHANNEL_STATE_STREAMING,
		VIDEO_CHANNEL_STATE_RECORDING,
		VIDEO_CHANNEL_STATE_IDLE,
		VIDEO_CHANNEL_STATE_UNKNOWN,
	}
}

// GetMetadata implements VideoChannelObject.
func (v *videoChannelObject) GetMetadata() ObjectMetadata {
	v.metadata.Type = "video_channel"
	return v.metadata
}

// RunAction implements VideoChannelObject.
func (v *videoChannelObject) RunAction(id, action string, payload []byte) error {

	switch action {
	case VIDEO_CHANNEL_ACTION_SNAPSHOT:
		var p SnapshotActionPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		r, err := v.snapshotFn(v, v.controller, p)
		if err != nil {
			v.controller.UpdateResultAttributes(id, map[string]string{"error": err.Error()})
			return err
		}
		return v.controller.UpdateResultAttributes(id, map[string]string{"filename": r})

	case VIDEO_CHANNEL_ACTION_VIDEOCLIP:
		var p VideoClipActionPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		r, err := v.videoclipFn(v, v.controller, p)
		if err != nil {
			v.controller.UpdateResultAttributes(id, map[string]string{"error": err.Error()})
			return err
		}
		return v.controller.UpdateResultAttributes(id, map[string]string{"filename": r})

	case VIDEO_CHANNEL_ACTION_PTZ_CONTROL:
		var p VideoChannelActionPtzControlPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return err
		}
		return v.ptzFn(v, v.controller, p)
	}
	return nil
}

// Setup implements VideoChannelObject.
func (v *videoChannelObject) Setup(oc ObjectController) error {
	v.controller = oc

	v.UpdateStateAttributes(map[string]string{
		"video_engine_id": v.videoEngineId,
		"stream_id":       v.streamId,
		"sub_stream_id":   v.subStreamId,
		"ptz":             strconv.FormatBool(v.ptz),
	})

	if v.setupFn != nil {
		return v.setupFn(v, oc)
	}
	return nil

}

type NewVideoChannelObjectProps struct {
	Metadata ObjectMetadata

	StreamID    string
	SubstreamID string
	VideoEngine string
	PTZ         bool
	Recording   bool

	SetupFn     func(VideoChannelObject, ObjectController) error
	SnapshotFn  func(VideoChannelObject, ObjectController, SnapshotActionPayload) (string, error)
	VideoclipFn func(VideoChannelObject, ObjectController, VideoClipActionPayload) (string, error)
	PtzFn       func(VideoChannelObject, ObjectController, VideoChannelActionPtzControlPayload) error
}

func NewVideoChannelObject(props NewVideoChannelObjectProps) VideoChannelObject {
	return &videoChannelObject{
		metadata:      props.Metadata,
		streamId:      props.StreamID,
		subStreamId:   props.SubstreamID,
		ptz:           props.PTZ,
		videoEngineId: props.VideoEngine,
		setupFn:       props.SetupFn,
		snapshotFn:    props.SnapshotFn,
		videoclipFn:   props.VideoclipFn,
		ptzFn:         props.PtzFn,
	}
}
