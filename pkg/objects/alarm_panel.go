package objects

const ALARM_PANEL_STATE_ARMED = "alarm_panel.state.armed"
const ALARM_PANEL_STATE_DISARMED = "alarm_panel.state.disarmed"

type AlarmPanelObject interface {
	RegistrableObject
	Arm(code string)
	Disarm(code string)
	Bypass(code string, zone string)
}

type alarmPanelObject struct{}

// Arm implements AlarmPanelObject.
func (a *alarmPanelObject) Arm(code string) {
	panic("unimplemented")
}

// Bypass implements AlarmPanelObject.
func (a *alarmPanelObject) Bypass(code string, zone string) {
	panic("unimplemented")
}

// Disarm implements AlarmPanelObject.
func (a *alarmPanelObject) Disarm(code string) {
	panic("unimplemented")
}

// GetAvailableActions implements AlarmPanelObject.
func (a *alarmPanelObject) GetAvailableActions() []ObjectAction {
	panic("unimplemented")
}

// GetAvailableStates implements AlarmPanelObject.
func (a *alarmPanelObject) GetAvailableStates() []string {
	panic("unimplemented")
}

// GetMetadata implements AlarmPanelObject.
func (a *alarmPanelObject) GetMetadata() ObjectMetadata {
	panic("unimplemented")
}

// RunAction implements AlarmPanelObject.
func (a *alarmPanelObject) RunAction(action string, payload []byte) error {
	panic("unimplemented")
}

// Setup implements AlarmPanelObject.
func (a *alarmPanelObject) Setup(ObjectController) error {
	panic("unimplemented")
}

type NewAlarmPanelObjectProps struct {
	// CodeIsRequired indicates if a code is required to arm/disarm the alarm panel. For frontend show a keypad.
	CodeIsRequired bool
	// CodeIsNumeric indicates if the code is numeric. For frontend show a numeric keypad.
	CodeIsNumeric bool
}

func NewAlarmPanelObject(props NewAlarmPanelObjectProps) AlarmPanelObject {
	return &alarmPanelObject{}
}
