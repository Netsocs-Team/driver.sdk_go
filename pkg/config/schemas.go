package config

type GetChannelResponseItem struct {
	ChannelNumber string `json:"channelNumber"`
	ChannelName   string `json:"name"`
	Source        string `json:"rtspSource"`
}

type GetChannelResponse []*GetChannelResponseItem

type SetUserRequest struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PrevPassword string `json:"prev_password"`
	NewPassword  string `json:"new_password"`
}

// https://github.com/Netsocs-Team/DevDocs/blob/main/markdown/drivers/config-schemas/setUsers.md
type SetUserResponse error

// https://github.com/Netsocs-Team/DevDocs/blob/main/markdown/drivers/config-schemas/deleteUser.md
type DeleteUserRequest struct {
	ID int `json:"id"`
}

// https://github.com/Netsocs-Team/DevDocs/blob/main/markdown/drivers/config-schemas/deleteUser.md
type DeleteUserResponse error

// https://github.com/Netsocs-Team/DevDocs/blob/main/markdown/drivers/config-schemas/getUsers.md
type GetUserRequest struct{}

// https://github.com/Netsocs-Team/DevDocs/blob/main/markdown/drivers/config-schemas/getUsers.md

type GetUserResponseItem struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	// 0: Online, 1: Offline, 2: Unknown
	Status int `json:"status"`
	// Created at in format ISO 8601
	CreatedAt *string `json:"created_at"`
	// Last login in format ISO 8601
	LastLogin *string `json:"last_login"`
	// Last logout in format ISO 8601
	LastLogout  *string `json:"last_logout"`
	LastIP      *string `json:"last_ip"`
	IsAdminUser bool    `json:"is_admin_user"`
}
type GetUserResponse []*GetUserResponseItem

// https://github.com/Netsocs-Team/DevDocs/blob/main/markdown/drivers/config-schemas/setAddPersonToAC.md
type SetAddPersonToACRequest struct {
	PersonID   string `json:"personId"`
	PersonName string `json:"personName"`
}

type SetAddPersonToACResponse error

// https://github.com/Netsocs-Team/DevDocs/blob/main/markdown/drivers/config-schemas/setCardToPersonAC.md
type SetCardToPersonACRequest struct {
	PersonID string   `json:"personId"`
	Cards    []string `json:"cards"`
}

type SetCardToPersonACResponse error

// https://github.com/Netsocs-Team/DevDocs/blob/main/markdown/drivers/config-schemas/setFaceToPersonAC.md
type SetFaceToPersonACRequest struct {
	PersonID string   `json:"personId"`
	Faces    []string `json:"faces"`
}

type SetFaceToPersonACResponse error

// https://github.com/Netsocs-Team/DevDocs/blob/main/markdown/drivers/config-schemas/setDelPersonToAC.md

type SetDelToPersonToACRequest struct {
	PersonID string `json:"personId"`
}

type SetDelToPersonToACResponse error

// https://github.com/Netsocs-Team/DevDocs/blob/main/markdown/drivers/config-schemas/getAdminUser.md

type GetAdminUserRequest struct{}

type GetAdminUserResponse []*GetAdminUserResponseItem

type GetAdminUserResponseItem struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`
}

type GetRecordingSourceRequest struct {
	ChannelNumber string  `json:"channelNumber"`
	UTCStart      string  `json:"utcStart"`
	UTCEnd        *string `json:"utcEnd"`
}

type GetRecordingSourceResponse struct {
	Source string `json:"rtspSource"`
}

type GetRecordingRangesRequest struct {
	ChannelNumber string  `json:"channelNumber"`
	UTCStart      *string `json:"utcStart"`
	UTCEnd        *string `json:"utcEnd"`
}

type RecordingRangeItem struct {
	UTCStart string `json:"utcStart"`
	UTCEnd   string `json:"utcEnd"`
}

type GetRecordingRangesResponse []*RecordingRangeItem

type ZoomDirection int

const (
	ZoomIn  ZoomDirection = 0
	ZoomOut ZoomDirection = 1
)

type ActionZoomRequest struct {
	ChannelNumber string        `json:"channelNumber"`
	ZoomDirection ZoomDirection `json:"zoomDirection"` // 0: zoom in, 1: zoom out
}

type ActionZoomResponse error

type ActionMoveRequestDirection int

const (
	MoveUp    ActionMoveRequestDirection = 0
	MoveDown  ActionMoveRequestDirection = 1
	MoveLeft  ActionMoveRequestDirection = 2
	MoveRight ActionMoveRequestDirection = 3
)

type ActionMoveRequest struct {
	ChannelNumber string                     `json:"channelNumber"`
	Direction     ActionMoveRequestDirection `json:"direction"`
}

type ActionMoveResponse error

type GetAvailableOutputsRequest struct{}

type GetAvailableOutputsResponse []*GetAvailableOutputsResponseItem
type GetAvailableOutputsResponseItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// 0 = relay, 1 = led
	Type int `json:"type"`
	// 0 = normalmente abierto, 1 = normalmente cerrado, 2 = encendido, 3 = apagado
	State int `json:"state"`
}

type GetAvailableSpeakersRequest struct{}
type GetAvailableSpeakersResponse []*GetAvailableSpeakersResponseItem
type GetAvailableSpeakersResponseItem struct {
	ID       string `json:"id"`
	Level    int    `json:"level"`
	MinLevel int    `json:"minLevel"`
	MaxLevel int    `json:"maxLevel"`
	Enabled  bool   `json:"enabled"`
	Name     string `json:"name"`
}

type GetStorageRequest struct{}
type GetStorageResponse []*GetStorageResponseItem
type GetStorageResponseItem struct {
	ID         string `json:"id"`
	DeviceName string `json:"deviceName"`
	Recording  bool   `json:"recording"`
	FreeSize   int    `json:"freeSize"`
	TotalSize  int    `json:"totalSize"`
	Status     int    `json:"status"`
}

type GetCurrentVideoResolutionByChannelRequest struct {
	ChannelNumber string `json:"channelNumber"`
}
type GetCurrentVideoResolutionByChannelResponse struct {
	Resolution string `json:"resolution"`
}
type GetAvailableVideoResolutionsRequest struct {
	ChannelNumber string `json:"channelNumber"`
}

type GetAvailableVideoResolutionsResponse struct {
	Resolutions []string `json:"resolutions"`
}

type SetVideoResolutionRequest struct {
	ChannelNumber string `json:"channelNumber"`
	Resolution    string `json:"resolution"`
}

type SetVideoResolutionResponse error

type SetVideoInBlackAndWhiteRequest struct {
	ChannelNumber string `json:"channelNumber"`
	Enabled       bool   `json:"enabled"`
}

type SetVideoInBlackAndWhiteResponse error

type GetVideoInBlackAndWhiteStatusRequest struct {
	ChannelNumber string `json:"channelNumber"`
}

type GetVideoInBlackAndWhiteStatusResponse struct {
	Enabled bool `json:"enabled"`
}

type SetMirrorVideoRequest struct {
	ChannelNumber string `json:"channelNumber"`
	Enabled       bool   `json:"enabled"`
}

type SetMirrorVideoResponse error

type GetMirrorVideoStatusRequest struct {
	ChannelNumber string `json:"channelNumber"`
}

type GetMirrorVideoStatusResponse struct {
	Enabled bool `json:"enabled"`
}

type SetFlipVideoRequest struct {
	ChannelNumber string `json:"channelNumber"`
	Enabled       bool   `json:"enabled"`
}

type SetFlipVideoResponse error

type GetFlipVideoStatusRequest struct {
	ChannelNumber string `json:"channelNumber"`
}

type GetFlipVideoStatusResponse struct {
	Enabled bool `json:"enabled"`
}

type ActionPlayAudioClipRequest struct {
	AudioClipURL string `json:"audioClipUrl"` // URL of the audio clip
	SpeakerID    string `json:"speakerId"`
}

type ActionPlayAudioClipResponse error

type GetHeatmapImageRequest struct {
	ChannelNumber string `json:"channelNumber"`
	// Start in format ISO 8601
	Start string `json:"start"`
	// End in format ISO 8601
	End string `json:"end"`
}

type GetHeatmapImageResponse struct {
	Filename string `json:"filename"`
}

type SetBlockPersonToACRequest struct {
	PersonID string `json:"personId"`
}

type SetBlockPersonToACResponse error

type DeleteAllPeopleACRequest struct{}

type DeleteAllPeopleACResponse error
type GetInputsResponseItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
	Type        int    `json:"type"`
}

// https://github.com/Netsocs-Team/DevDocs/blob/main/markdown/drivers/config-schemas/getChannels.md
type GetInputsResponse []*GetInputsResponseItem

type GetMicrophoneResponseItem struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	ID      string `json:"id"`
}

type GetMicrophoneResponse []*GetMicrophoneResponseItem

type GetFtpInfoRequest struct{}

type GetFtpInfoResponseItem struct {
	ID         string `json:"id"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Directory  string `json:"directory"`
	Encryption string `json:"encryption"`
	Mode       string `json:"mode"`
}
type GetFtpInfoResponse []*GetFtpInfoResponseItem

type SetFtpInfoRequest struct {
	ID         string `json:"id"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Directory  string `json:"directory"`
	Encryption string `json:"encryption"`
	Mode       string `json:"mode"`
}

type SetFtpInfoResponse error

type GetAllPeopleFromACRequest struct{}

type GetAllPeopleFromACResponseItem struct {
	PersonID              string `json:"personId"`
	PersonName            string `json:"personName"`
	TotalCards            int    `json:"totalCards"`
	TotalFaces            int    `json:"totalFaces"`
	TotalFingerprints     int    `json:"totalFingerprints"`
	TotalIris             int    `json:"totalIris"`
	TotalVoiceRecognition int    `json:"totalVoiceRecognition"`
	TotalRFID             int    `json:"totalRFID"`
}

type GetAllPeopleFromACResponse []*GetAllPeopleFromACResponseItem

type GetAlarmPartitionsResponseItem struct {
	PartitionId  string `json:"partitionId"`
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	ArmedStatus  int    `json:"armedStatus"`
	TroubleState int    `json:"troubleState"`
	AlarmState   int    `json:"alarmState"`
}

type GetAlarmPartitionsResponse []*GetAlarmPartitionsResponseItem

type SetAlarmPartitionRequest struct {
	PartitionId string `json:"partitionId"`
	Name        string `json:"name"`
}

type SetAlarmPartitionResponse error

type GetAlarmZonesResponseItem struct {
	ZoneId     string `json:"zoneId"`
	Name       string `json:"name"`
	AlarmState int    `json:"alarmState"`
	Bypassed   bool   `json:"bypassed"`
	OpenClosed bool   `json:"openClosed"`
}

type GetAlarmZonesResponse []*GetAlarmZonesResponseItem

type SetAlarmZoneRequest struct {
	PartitionId string `json:"partitionId"`
	ZoneId      string `json:"zoneId"`
	Name        string `json:"name"`
}

type SetAlarmZoneResponse error

type SetAlarmPartitionZoneBypassRequest struct {
	PartitionId string `json:"partitionId"`
	ZoneId      string `json:"zoneId"`
	Bypass      bool   `json:"bypass"`
}

type SetAlarmPartitionZoneBypassResponse error

type GetAlarmZoneStatusRequest struct {
	ZoneId      string `json:"zoneId"`
	PartitionId string `json:"partitionId"`
}

type GetAlarmZoneStatusResponse struct {
	AlarmState int  `json:"alarmState"`
	Bypassed   bool `json:"bypassed"`
	OpenClosed bool `json:"openClosed"`
}

type SetAddAlarmUserRequest struct {
	PartitionId string `json:"partitionId"`
	UserName    string `json:"userName"`
	AccessCode  string `json:"accessCode"`
}

type SetAddAlarmUserResponse error

type SetAlarmUserRequest struct {
	UserId      string `json:"userId"`
	PartitionId string `json:"partitionId"`
	UserName    string `json:"userName"`
	AccessCode  string `json:"accessCode"`
}

type SetAlarmUserResponse error

type GetAlarmUsersResponseItem struct {
	UserId      string `json:"userId"`
	PartitionId string `json:"partitionId"`
	UserName    string `json:"userName"`
}

type GetAlarmUsersResponse []*GetAlarmUsersResponseItem

type GetAlarmArmStatesResponseItem struct {
	ArmedStatus string `json:"armedStatus"`
	Value       int    `json:"value"`
}

type GetAlarmArmStatesResponse []*GetAlarmArmStatesResponseItem

type GetFAPStatesResponseItem struct {
	FAPStatus string `json:"FAPStatus"`
	Value     int    `json:"value"`
}

type GetAlarmFAPStatesResponse []*GetFAPStatesResponseItem

type ActionAlarmArmPartitionRequest struct {
	PartitionId string `json:"partitionId"`
	Value       int    `json:"value"`
}

type ActionAlarmArmPartitionResponse error

type ActionAlarmDisarmPartitionRequest struct {
	PartitionId string `json:"partitionId"`
}

type ActionAlarmDisarmPartitionResponse error

type ActionAlarmFAPPartitionRequest struct {
	PartitionId string `json:"partitionId"`
	Value       int    `json:"value"`
}

type ActionAlarmFAPPartitionResponse error

type GetAlarmPartitionZonesRequest struct {
	PartitionId string `json:"partitionId"`
}

type GetAlarmPartitionZonesResponse []*GetAlarmZonesResponseItem

type SetAddAlarmPartitionZoneRequest struct {
	PartitionId string `json:"partitionId"`
	ZoneId      string `json:"zoneId"`
}

type SetAddAlarmPartitionZoneResponse error

type SetDeleteAlarmPartitionZoneRequest struct {
	PartitionId string `json:"partitionId"`
	ZoneId      string `json:"zoneId"`
}

type SetDeleteAlarmPartitionZoneResponse error

type SetQRToPersonACRequest struct {
	PersonID string                         `json:"personId"`
	QRCodes  []SetQRToPersonACRequestQRCode `json:"qrcodes"`
}

type SetQRToPersonACRequestQRCode struct {
	Value string `json:"value"`
	Type  QRType `json:"type"`
}

type QRType int16

const (
	QR_DEFAULT QRType = 50
)

type SetQRToPersonACResponse error

type SetBackgroundImageRequest struct {
	ImageURL string `json:"imageUrl"`
}

type SetBackgroundImageResponse error

type SetActionUnlockDeviceRequest struct {
	DeviceID      string `json:"deviceId"`
	Unlock        bool   `json:"unlock"`
	UnlockCode    string `json:"unlockCode"`
	UnlockMode    string `json:"unlockMode"`
	AditionalData string `json:"aditionalData"`
}

type SetActionUnlockDeviceResponse error

type GetUnlockDeviceStatusRequest struct {
	DeviceID string `json:"deviceId"`
}

type GetUnlockDeviceStatusResponse struct {
	UnlockStatus string `json:"unlockStatus"`
}

type GetSubdevicesRequest struct{}

type GetSubdevicesResponseItem struct {
	IDBrand          int    `json:"id_brand"`
	IDModel          int    `json:"id_model"`
	IDManufacturer   int    `json:"id_manufacturer"`
	IDSubSystem      int    `json:"id_sub_system"`
	IDDeviceGroup    int    `json:"id_device_group"`
	ChildID          string `json:"child_id"`
	ChildName        string `json:"child_name"`
	ChildDescription string `json:"child_description"`
	ChildIP          string `json:"child_ip"`
	ChildPort        int    `json:"child_port"`
}

type GetSubdevicesResponse []GetSubdevicesResponseItem

type ExtraFieldType string

const (
	ExtraFieldTypeString  ExtraFieldType = "string"
	ExtraFieldTypeInt     ExtraFieldType = "int"
	ExtraFieldTypeFloat   ExtraFieldType = "float"
	ExtraFieldTypeBoolean ExtraFieldType = "boolean"
)

type ExtraDeviceFieldsDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Type        ExtraFieldType `json:"type"`
}

type GetDeviceExtraFieldsResponse = []*ExtraDeviceFieldsDefinition

type GetEventsAvailableRequest struct{}
type EventSchemaTranslationStrings struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type GetEventsAvailableResponseItem struct {
	// Key is unique identifier for the event
	// its value is created by the concatenation of the device name and the event name
	// Example: "driver.tyco_entrapass:identificationSuccess"
	Key string `json:"key"`
	// Name and description are the human readable values for the event
	TranslationStrings map[string]EventSchemaTranslationStrings `json:"translation_strings"`
	// Value between 1 - 3. 1 -> low, 2 -> medium, 3 -> high
	Priority int `json:"priority"`
}

type GetEventsAvailableResponse []GetEventsAvailableResponseItem

type GetPeopleCountingRequest struct {
	ChannelNumber string `json:"channelNumber"`
	UTCStart      string `json:"utcStart"`
	UTCEnd        string `json:"utcEnd"`
}

type GetPeopleCountingResponseItem struct {
	LineID   string `json:"lineId"`
	Name     string `json:"name"`
	InCount  int    `json:"inCount"`
	OutCount int    `json:"outCount"`
}

type GetPeopleCountingResponse []GetPeopleCountingResponseItem
