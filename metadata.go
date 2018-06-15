package airvantage

import (
	"encoding/json"
)

// Metadata commonly found in other struct.
type Metadata map[string]string

type jsonMetadata struct {
	Key   string
	Value string
}

func (m *Metadata) UnmarshalJSON(b []byte) error {
	var array []jsonMetadata
	if err := json.Unmarshal(b, &array); err != nil {
		return err
	}

	h := make(map[string]string)

	for _, meta := range array {
		h[meta.Key] = meta.Value
	}

	*m = h

	return nil
}

func (m Metadata) MarshalJSON() ([]byte, error) {
	var array []jsonMetadata

	for k, v := range m {
		array = append(array, jsonMetadata{Key: k, Value: v})
	}

	return json.Marshal(array)
}
