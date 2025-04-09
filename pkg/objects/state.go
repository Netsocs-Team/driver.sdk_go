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

type ObjectStateChange struct {
	ObjectID                  string            `json:"id"`
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
