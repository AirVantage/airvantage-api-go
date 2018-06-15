package airvantage

import (
	"encoding/json"
	"time"
)

// AVTime is a Unix timestamp in milliseconds.
type AVTime int64

func (t *AVTime) UnmarshalJSON(b []byte) error {
	var stamp int64
	if err := json.Unmarshal(b, &stamp); err != nil {
		return err
	}

	*t = AVTime(stamp)

	return nil
}

func (t AVTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(t))
}

// NewAVTime creates an AVTime from a Go Time struct.
func NewAVTime(t time.Time) AVTime {
	return AVTime(t.UnixNano() / 1000)
}

// Time converts an AVTime to a Go Time struct.
func (t AVTime) Time() time.Time {
	return time.Unix(int64(t)/1000, int64(t)%1000*int64(time.Millisecond))
}
