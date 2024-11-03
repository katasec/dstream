package cdc

import (
	"encoding/json"
)

// JsonEvent represents the simplified JSON structure for CDC events
type JsonEvent struct {
	Table string                 `json:"table"`
	Op    string                 `json:"op"`
	Data  map[string]interface{} `json:"data,omitempty"`
	LSN   string                 `json:"lsn"`
}

// toJsonPayload formats the CDC event as a JsonEvent instance
func (j *JsonEvent) toJsonPayload() ([]byte, error) {
	return json.MarshalIndent(j, "", "    ")
}
