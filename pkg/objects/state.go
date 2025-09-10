package objects

type StateRecord struct {
	ID       string `json:"id"`
	ObjectID string `json:"object_id" bson:"object_id"`
	// "2019-07-30T06:43:40.252Z" - ISO 8601
	Datetime string `json:"datetime"`
	Domain   string `json:"domain"`
	State    State  `json:"state"`
}
type State struct {
	State                     string            `json:"state"`
	StateAdditionalProperties map[string]string `json:"state_additional_properties" bson:"state_additional_properties"`
}

type ChangeStateBatchRequest struct {
	Changes []ObjectStateChange `json:"changes"`
}

type ChangeStateBatchResponse struct {
	ID          string                      `json:"object_id"`
	Datetime    string                      `json:"datetime"`
	Error       string                      `json:"error"`
	Changed     bool                        `json:"changed"`
	ObjectState ObjectStateChangedEventData `json:"object_state"`
}

type ObjectStateChangedEventData struct {
	ID        string `json:"id"`
	ObjectID  string `json:"object_id"`
	Datetime  string `json:"datetime"`
	Domain    string `json:"domain"`
	PrevState struct {
		State                     string            `json:"state"`
		StateAdditionalProperties map[string]string `json:"state_additional_properties"`
	} `json:"prev_state"`
	NewState struct {
		State                     string            `json:"state"`
		StateAdditionalProperties map[string]string `json:"state_additional_properties"`
	} `json:"new_state"`
}

type ObjectStateChange struct {
	ObjectID                  string            `json:"object_id"`
	State                     string            `json:"state"`
	StateAdditionalProperties map[string]string `json:"state_additional_properties" bson:"state_additional_properties"`
}

type PaginatedStateRecord struct {
	Items    []StateRecord `json:"items"`
	Metadata struct {
		TotalItems int `json:"total_items"`
		Limit      int `json:"limit"`
		Offset     int `json:"offset"`
	} `json:"_metadata"`
}
