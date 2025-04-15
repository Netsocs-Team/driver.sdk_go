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
