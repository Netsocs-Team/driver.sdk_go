package objects

type Event struct {
	ObjectIDs  []string
	ImageURLs  []string
	VideoURLs  []string
	Properties map[string]string
}

type NewEventRequestBodySchema struct {
	EventType                 string            `json:"event_type"`
	Rels                      []string          `json:"rels"`
	EventAdditionalProperties map[string]string `json:"event_additional_properties"`
	Images                    []string          `json:"images"`
	VideoClips                []string          `json:"video_clips"`
}

type EventRecord struct {
	ID                        string            `json:"id,omitempty"`
	Report                    ReportRecord      `json:"report,omitempty"`
	Rels                      []string          `json:"rels,omitempty"`
	EventAdditionalProperties map[string]string `json:"event_additional_properties,omitempty"`
	Images                    []string          `json:"images,omitempty"`
	VideoClips                []string          `json:"video_clips,omitempty"`
	EventType                 string            `json:"event_type,omitempty"`
	CreatedAt                 string            `json:"created_at,omitempty"`
	UpdatedAt                 string            `json:"updated_at,omitempty"`
	Domain                    string            `json:"domain"`
}

type EventDispatchResponse struct {
	ID string `json:"id"`
}

type ReportRecord struct {
	WrittenBy string `json:"written_by"`
	WrittenAt string `json:"written_at"`
	Content   string `json:"content"`
}
