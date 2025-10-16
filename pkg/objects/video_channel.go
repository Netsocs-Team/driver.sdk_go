package objects

import (
	"fmt"
	"strconv"
	"time"

	"github.com/goccy/go-json"
)

const VIDEO_CHANNEL_STATE_STREAMING = "video_channel.state.streaming"
const VIDEO_CHANNEL_STATE_RECORDING = "video_channel.state.recording"
const VIDEO_CHANNEL_STATE_IDLE = "video_channel.state.idle"
const VIDEO_CHANNEL_STATE_UNKNOWN = "video_channel.state.unknown"

const VIDEO_CHANNEL_ACTION_SNAPSHOT = "video_channel.action.snapshot"
const VIDEO_CHANNEL_ACTION_VIDEOCLIP = "video_channel.action.videoclip"
const VIDEO_CHANNEL_ACTION_PTZ_CONTROL = "video_channel.action.ptz_control"
const VIDEO_CHANNEL_ACTION_SEEK = "video_channel.action.seek" //seek to a specific timestamp to playback id

// seek states
const VIDEO_CHANNEL_SEEK_STATE_MEDIA_NOT_FOUND = "media_not_found"
const VIDEO_CHANNEL_SEEK_STATE_BAD_STATUS_CODE = "bad_status_code"
const VIDEO_CHANNEL_SEEK_STATE_PLAYING = "playing"
const VIDEO_CHANNEL_SEEK_STATE_SEEKING = "seeking"
const VIDEO_CHANNEL_SEEK_STATE_VIDEO_ENGINE_NOT_AVAILABLE = "video_engine_not_available"

type PTZCommand string

const (
	PTZ_COMMAND_UP         PTZCommand = "up"
	PTZ_COMMAND_DOWN       PTZCommand = "down"
	PTZ_COMMAND_LEFT       PTZCommand = "left"
	PTZ_COMMAND_RIGHT      PTZCommand = "right"
	PTZ_COMMAND_UP_LEFT    PTZCommand = "up_left"
	PTZ_COMMAND_UP_RIGHT   PTZCommand = "up_right"
	PTZ_COMMAND_DOWN_LEFT  PTZCommand = "down_left"
	PTZ_COMMAND_DOWN_RIGHT PTZCommand = "down_right"
	PTZ_COMMAND_ZOOM_IN    PTZCommand = "zoom_in"
	PTZ_COMMAND_ZOOM_OUT   PTZCommand = "zoom_out"
	PTZ_COMMAND_STOP       PTZCommand = "stop"
	PTZ_COMMAND_FOCUS_NEAR PTZCommand = "focus_near"
	PTZ_COMMAND_FOCUS_FAR  PTZCommand = "focus_far"
	PTZ_COMMAND_IRIS_OPEN  PTZCommand = "iris_open"
	PTZ_COMMAND_IRIS_CLOSE PTZCommand = "iris_close"
)

const PTZ_MAX_SPEED = 10
const PTZ_MIN_SPEED = 1

type VideoChannelActionPtzControlPayload struct {
	Command  PTZCommand `json:"command"` //up, down, left, right, zoom_in, zoom_out, stop
	Value    int        `json:"value"`   //speed value from 1 to 10
	Relative bool       `json:"relative"`
}

type VideoClipActionPayload struct {
	StartTimestamp string `json:"start_timestamp"`
	EndTimestamp   string `json:"end_timestamp"`
	Resolution     string `json:"resolution"` //"1920x1080"
	Timeout        int    `json:"timeout"`    // This is to stop trying to make the video clip after certain seconds
}

type SnapshotActionPayload struct {
	Timestamp  string `json:"snapshot_timestamp"` //if its empty make it as soon as received
	Resolution string `json:"resolution"`         //"1920x1080"
}

type VideoChannelActionPtzControlPayloadDirection string

// Metadata representa el mensaje completo de metadatos
type AnalyticAnnotations struct {
	Timestamp time.Time `json:"timestamp"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`

	Objects     []Object     `json:"objects,omitempty"`
	Annotations []Annotation `json:"annotations,omitempty"`
}

// Object representa un objeto detectado
type Object struct {
	ID         string  `json:"id"`
	Type       string  `json:"type"` // person, vehicle, face, etc.
	Confidence float64 `json:"confidence"`

	// Bounding box
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`

	// Atributos opcionales
	Attributes map[string]string `json:"attributes,omitempty"`
	TrackID    string            `json:"track_id,omitempty"`

	// Estilo de dibujo
	Color     string `json:"color,omitempty"`      // Color del borde
	FillColor string `json:"fill_color,omitempty"` // Color de relleno
	LineWidth int    `json:"line_width,omitempty"`
}

// Annotation representa una figura o texto para dibujar
type Annotation struct {
	Type string `json:"type"` // rectangle, polygon, line, circle, text

	// Para todas las figuras
	Points []Point `json:"points,omitempty"` // Para polygon, line

	// Para rectangle y circle
	X      int `json:"x,omitempty"`
	Y      int `json:"y,omitempty"`
	Width  int `json:"width,omitempty"`
	Height int `json:"height,omitempty"`
	Radius int `json:"radius,omitempty"` // Para circle

	// Para text
	Text     string `json:"text,omitempty"`
	FontSize int    `json:"font_size,omitempty"`

	// Estilo
	Color     string `json:"color,omitempty"`
	FillColor string `json:"fill_color,omitempty"`
	LineWidth int    `json:"line_width,omitempty"`

	Label string `json:"label,omitempty"`
}

// Point representa un punto 2D
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

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
	SetAnalyticsMetadata(metadata AnalyticAnnotations) error
}

type SeekPayload struct {
	PlaybackID string  `json:"playback_id"` //playback id to seek to
	SeekTo     string  `json:"seek_to"`     //time to seek to
	Speed      float32 `json:"speed"`       //speed to play the video, 1.0 is normal speed, 2.0 is double speed, 0.5 is half speed
	Reverse    bool    `json:"reverse"`     //if true, play the video in reverse
	Destroy    bool    `json:"destroy"`     //if true, destroy the playback after seeking
	// video_engine_hostname
	VideoEngineHostname string `json:"video_engine_hostname"`
	// video_engine_rtsp_port
	VideoEngineRtspPort string `json:"video_engine_rtsp_port"`
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
	seekFn      func(VideoChannelObject, ObjectController, SeekPayload) error
}

// SetAnalyticsMetadata implements VideoChannelObject.
func (v *videoChannelObject) SetAnalyticsMetadata(metadata AnalyticAnnotations) error {
	json, err := json.Marshal(metadata)
	if err != nil {
		return err
	}
	return v.controller.UpdateStateAttributes(v.GetMetadata().ObjectID, map[string]string{"analytics_metadata": string(json)})
}

// UpdateStateAttributes implements VideoChannelObject.
func (v *videoChannelObject) UpdateStateAttributes(attributes map[string]string) error {
	return v.controller.UpdateStateAttributes(v.GetMetadata().ObjectID, attributes)
}

// SetState implements VideoChannelObject.
func (v *videoChannelObject) SetState(state string) error {
	return v.controller.SetState(v.GetMetadata().ObjectID, state)
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
		{Action: VIDEO_CHANNEL_ACTION_SEEK, Domain: v.metadata.Domain},
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
func (v *videoChannelObject) RunAction(id, action string, payload []byte) (map[string]string, error) {

	switch action {
	case VIDEO_CHANNEL_ACTION_SNAPSHOT:
		var p SnapshotActionPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return nil, err
		}
		r, err := v.snapshotFn(v, v.controller, p)
		if err != nil {

			return nil, err
		}
		return map[string]string{"snapshot_link": r}, nil

	case VIDEO_CHANNEL_ACTION_VIDEOCLIP:
		var p VideoClipActionPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return nil, err
		}
		r, err := v.videoclipFn(v, v.controller, p)
		if err != nil {
			return nil, err
		}
		return map[string]string{"videoclip_link": r}, nil

	case VIDEO_CHANNEL_ACTION_PTZ_CONTROL:
		var p VideoChannelActionPtzControlPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return nil, err
		}
		return nil, v.ptzFn(v, v.controller, p)

	case VIDEO_CHANNEL_ACTION_SEEK:
		var p SeekPayload
		if err := json.Unmarshal(payload, &p); err != nil {
			return nil, err
		}
		return nil, v.seekFn(v, v.controller, p)
	}

	return nil, fmt.Errorf("action %s not found", action)

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
	SeekFn      func(VideoChannelObject, ObjectController, SeekPayload) error
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
		seekFn:        props.SeekFn,
	}
}
