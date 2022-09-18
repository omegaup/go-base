package base

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"unicode"
)

// Duration is identical to time.Duration, except it can implements the
// json.Marshaler interface with time.Duration.String() and
// time.Duration.ParseDuration().
type Duration time.Duration

// String returns a string representing the duration in the form "72h3m0.5s".
// Leading zero units are omitted. As a special case, durations less than one
// second format use a smaller unit (milli-, micro-, or nanoseconds) to ensure
// that the leading digit is non-zero. The zero duration formats as 0s.
func (d Duration) String() string {
	return time.Duration(d).String()
}

// MarshalJSON implements the json.Marshaler interface. The duration is a quoted
// string in RFC 3339 format, with sub-second precision added if present.
func (d Duration) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%s\"", d.String())), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface. The duration is
// expected to be a quoted string that time.ParseDuration() can understand.
func (d *Duration) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	if unicode.IsDigit(rune(data[0])) {
		val, err := strconv.ParseFloat(string(data), 64)
		if err != nil {
			return nil
		}
		*d = Duration(time.Duration(val*1e9) * time.Nanosecond)
		return nil
	}
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return errors.New("time: invalid duration " + string(data))
	}
	parsed, err := time.ParseDuration(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	*d = Duration(parsed)
	return nil
}

// Milliseconds returns the duration as a floating point number of milliseconds.
func (d Duration) Milliseconds() float64 {
	return float64(d) / float64(time.Millisecond)
}

// Seconds returns the duration as a floating point number of seconds.
func (d Duration) Seconds() float64 {
	return time.Duration(d).Seconds()
}
