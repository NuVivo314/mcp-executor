package sandbox

import (
	"encoding/json"
	"time"
)

// jsonMarshal serialises v to a JSON string. Returns "null" on error.
func jsonMarshal(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(b)
}

// jsonUnmarshal deserialises JSON into v.
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// deadlineNow returns the current time (extracted for testability).
var deadlineNow = time.Now
