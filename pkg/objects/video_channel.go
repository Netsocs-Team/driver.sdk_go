package objects

const VIDEO_CHANNEL_STATE_STREAMING = "video_channel.state.streaming"
const VIDEO_CHANNEL_STATE_RECORDING = "video_channel.state.recording"
const VIDEO_CHANNEL_STATE_IDLE = "video_channel.state.idle"
const VIDEO_CHANNEL_STATE_UNKNOWN = "video_channel.state.unknown"

const VIDEO_CHANNEL_ACTION_SNAPSHOT = "video_channel.action.snapshot"
const VIDEO_CHANNEL_ACTION_PTZ_CONTROL = "video_channel.action.ptz_control"

type VideoChannelActionPtzControlPayload struct {
	Pan      int  `json:"pan"`
	Tilt     int  `json:"tilt"`
	Zoom     int  `json:"zoom"`
	Relative bool `json:"relative"`
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
	setup      func(VideoChannelObject, ObjectController) error
	controller ObjectController
	metadata   ObjectMetadata
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
	return v.controller.UpdateStateAttributes(v.GetMetadata().ObjectID, map[string]interface{}{"primary_stream": streamId})
}

// SecundaryStream implements VideoChannelObject.
func (v *videoChannelObject) SecondaryStream(streamId string) error {
	return v.controller.UpdateStateAttributes(v.GetMetadata().ObjectID, map[string]interface{}{"secondary_stream": streamId})
}

// GetAvailableActions implements VideoChannelObject.
func (v *videoChannelObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{
		{Action: VIDEO_CHANNEL_ACTION_SNAPSHOT, Domain: v.metadata.Domain},
		{Action: VIDEO_CHANNEL_ACTION_PTZ_CONTROL, Domain: v.metadata.Domain},
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
	return v.metadata
}

// RunAction implements VideoChannelObject.
func (v *videoChannelObject) RunAction(action string, payload []byte) error {
	panic("unimplemented")
}

// Setup implements VideoChannelObject.
func (v *videoChannelObject) Setup(oc ObjectController) error {
	v.controller = oc
	return v.setup(v, oc)
}

func NewVideoChannelObject(metadata ObjectMetadata, setup func(VideoChannelObject, ObjectController) error) VideoChannelObject {
	return &videoChannelObject{
		metadata: metadata,
		setup:    setup,
	}
}
