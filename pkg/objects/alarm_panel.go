package objects

import (
	"fmt"
	"strings"

	"github.com/goccy/go-json"
)

const ALARM_PANEL_STATE_UNKNOWN = "alarm_panel.state.unknown"
const ALARM_PANEL_STATE_DISARMED = "alarm_panel.state.disarmed"
const ALARM_PANEL_STATE_STAY_ARMED = "alarm_panel.state.stay_armed"
const ALARM_PANEL_STATE_AWAY_ARMED = "alarm_panel.state.away_armed"
const ALARM_PANEL_STATE_STAY_ARMED_WITH_NO_ENTRY_DELAY = "alarm_panel.state.stay_armed_with_no_entry_delay"
const ALARM_PANEL_STATE_AWAY_ARMED_WITH_NO_ENTRY_DELAY = "alarm_panel.state.away_armed_with_no_entry_delay"
const ALARM_PANEL_STATE_NIGHT_MODE_ARMED = "alarm_panel.state.night_mode_armed"
const ALARM_PANEL_STATE_INTERIOR_ARMED = "alarm_panel.state.interior_armed"
const ALARM_PANEL_STATE_USER_ARMED = "alarm_panel.state.user_armed"
const ALARM_PANEL_STATE_ERROR = "alarm_panel.state.error"

const ALARM_PANEL_ACTION_ARM = "alarm_panel.action.arm"
const ALARM_PANEL_ACTION_DISARM = "alarm_panel.action.disarm"
const ALARM_PANEL_ACTION_FIRE = "alarm_panel.action.fire"
const ALARM_PANEL_ACTION_PANIC = "alarm_panel.action.panic"
const ALARM_PANEL_ACTION_AUXILIARY = "alarm_panel.action.auxiliary"
const ALARM_PANEL_BYPASS = "alarm_panel.bypass"

type AlarmPanelObject interface {
	RegistrableObject
	SetBypassedZones(zones []string) error
}

type actionPayload struct {
	Code       string `json:"code"`
	Zone       string `json:"zone"`
	BypassMode bool   `json:"bypass_mode"`
	ArmMode    string `json:"arm_mode"`
}
type alarmPanelObject struct {
	controller ObjectController
	metadata   ObjectMetadata

	armFn    func(alarmPanelObject AlarmPanelObject, oc ObjectController, mode string, key string) error
	disarmFn func(alarmPanelObject AlarmPanelObject, oc ObjectController, key string) error

	fireFn      func(alarmPanelObject AlarmPanelObject, oc ObjectController, key string) error
	panicFn     func(alarmPanelObject AlarmPanelObject, oc ObjectController, key string) error
	auxiliaryFn func(alarmPanelObject AlarmPanelObject, oc ObjectController, key string) error
	bypassFn    func(alarmPanelObject AlarmPanelObject, oc ObjectController, key string, zoneId string) error
	setupFn     func(alarmPanelObject AlarmPanelObject, oc ObjectController) error
}

// UpdateStateAttributes implements AlarmPanelObject.
func (a *alarmPanelObject) UpdateStateAttributes(attributes map[string]string) error {
	return a.controller.UpdateStateAttributes(a.GetMetadata().ObjectID, attributes)
}

// SetBypassedZones implements AlarmPanelObject.
func (a *alarmPanelObject) SetBypassedZones(zones []string) error {
	return a.controller.UpdateStateAttributes(a.GetMetadata().ObjectID, map[string]string{
		"bypassed_zones": strings.Join(zones, ","),
	})
}

// SetState implements AlarmPanelObject.
func (a *alarmPanelObject) SetState(state string) error {
	return a.controller.SetState(a.GetMetadata().ObjectID, state)
}

// GetAvailableActions implements AlarmPanelObject.
func (a *alarmPanelObject) GetAvailableActions() []ObjectAction {
	return []ObjectAction{
		{
			Action: ALARM_PANEL_ACTION_ARM,
			Domain: a.GetMetadata().Domain,
		},
		{
			Action: ALARM_PANEL_ACTION_DISARM,
			Domain: a.GetMetadata().Domain,
		},
	}
}

// GetAvailableStates implements AlarmPanelObject.
func (a *alarmPanelObject) GetAvailableStates() []string {
	return []string{
		ALARM_PANEL_STATE_UNKNOWN,
		ALARM_PANEL_STATE_DISARMED,
		ALARM_PANEL_STATE_STAY_ARMED,
		ALARM_PANEL_STATE_AWAY_ARMED,
		ALARM_PANEL_STATE_STAY_ARMED_WITH_NO_ENTRY_DELAY,
		ALARM_PANEL_STATE_AWAY_ARMED_WITH_NO_ENTRY_DELAY,
		ALARM_PANEL_STATE_NIGHT_MODE_ARMED,
		ALARM_PANEL_STATE_INTERIOR_ARMED,
		ALARM_PANEL_STATE_USER_ARMED,
		ALARM_PANEL_STATE_ERROR,
		ALARM_PANEL_BYPASS,
	}
}

// GetMetadata implements AlarmPanelObject.
func (a *alarmPanelObject) GetMetadata() ObjectMetadata {
	a.metadata.Type = "alarm_panel"
	return a.metadata
}

// RunAction implements AlarmPanelObject.
func (a *alarmPanelObject) RunAction(id, action string, payload []byte) (map[string]string, error) {

	var p actionPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return nil, err
	}

	switch action {
	case ALARM_PANEL_ACTION_ARM:
		if err := a.armFn(a, a.controller, p.ArmMode, p.Code); err != nil {
			return nil, err
		}
		return nil, a.controller.SetState(a.GetMetadata().ObjectID, ALARM_PANEL_STATE_AWAY_ARMED)
	case ALARM_PANEL_ACTION_DISARM:
		if err := a.disarmFn(a, a.controller, p.Code); err != nil {
			return nil, err
		}
		return nil, a.controller.SetState(a.GetMetadata().ObjectID, ALARM_PANEL_STATE_DISARMED)
	case ALARM_PANEL_ACTION_FIRE:
		return nil, a.fireFn(a, a.controller, p.Code)
	case ALARM_PANEL_ACTION_PANIC:
		return nil, a.panicFn(a, a.controller, p.Code)
	case ALARM_PANEL_ACTION_AUXILIARY:
		return nil, a.auxiliaryFn(a, a.controller, p.Code)
	case ALARM_PANEL_BYPASS:
		return nil, a.bypassFn(a, a.controller, p.Code, p.Zone)
	}
	return nil, fmt.Errorf("action %s not found", action)
}

// Setup implements AlarmPanelObject.
func (a *alarmPanelObject) Setup(oc ObjectController) error {
	a.controller = oc

	a.setupFn(a, oc)
	return nil
}

type NewAlarmPanelObjectProps struct {
	// CodeIsRequired indicates if a code is required to arm/disarm the alarm panel. For frontend show a keypad.
	CodeIsRequired bool
	// CodeIsNumeric indicates if the code is numeric. For frontend show a numeric keypad.
	CodeIsNumeric bool

	Metadata ObjectMetadata

	SetupFn  func(alarmPanelObject AlarmPanelObject, oc ObjectController) error
	ArmFn    func(alarmPanelObject AlarmPanelObject, oc ObjectController, mode string, key string) error
	DisarmFn func(alarmPanelObject AlarmPanelObject, oc ObjectController, key string) error

	FireFn      func(alarmPanelObject AlarmPanelObject, oc ObjectController, key string) error
	PanicFn     func(alarmPanelObject AlarmPanelObject, oc ObjectController, key string) error
	AuxiliaryFn func(alarmPanelObject AlarmPanelObject, oc ObjectController, key string) error
}

func NewAlarmPanelObject(props NewAlarmPanelObjectProps) AlarmPanelObject {
	return &alarmPanelObject{
		metadata:    props.Metadata,
		setupFn:     props.SetupFn,
		armFn:       props.ArmFn,
		disarmFn:    props.DisarmFn,
		fireFn:      props.FireFn,
		panicFn:     props.PanicFn,
		auxiliaryFn: props.AuxiliaryFn,
	}
}
