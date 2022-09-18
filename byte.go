package base

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// A Byte is a unit of digital information.
type Byte int64

const (
	// Kibibyte is 1024 Bytes.
	Kibibyte = Byte(1024)

	// Mebibyte is 1024 Kibibytes.
	Mebibyte = Byte(1024) * Kibibyte

	// Gibibyte is 1024 Mebibytes.
	Gibibyte = Byte(1024) * Mebibyte

	// Tebibyte is 1024 Gibibytes.
	Tebibyte = Byte(1024) * Gibibyte
)

// Bytes returns the Byte as an integer number of bytes.
func (b Byte) Bytes() int64 {
	return int64(b)
}

// Kibibytes returns the Byte as an floating point number of Kibibytes.
func (b Byte) Kibibytes() float64 {
	return float64(b) / float64(Kibibyte)
}

// Mebibytes returns the Byte as an floating point number of Mebibytes.
func (b Byte) Mebibytes() float64 {
	return float64(b) / float64(Mebibyte)
}

// Gibibytes returns the Byte as an floating point number of Gibibytes.
func (b Byte) Gibibytes() float64 {
	return float64(b) / float64(Gibibyte)
}

// Tebibytes returns the Byte as an floating point number of Tebibytes.
func (b Byte) Tebibytes() float64 {
	return float64(b) / float64(Tebibyte)
}

// MarshalJSON implements the json.Marshaler interface. The result is an
// integer number of bytes.
func (b Byte) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", b.Bytes())), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface. The result can be
// an integer number of bytes, or a quoted string that MarshalJSON() can
// understand.
func (b *Byte) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	if unicode.IsDigit(rune(data[0])) {
		val, err := strconv.ParseInt(string(data), 10, 64)
		if err != nil {
			return nil
		}
		*b = Byte(val)
		return nil
	}
	unquoted := string(data)
	if len(unquoted) < 3 || unquoted[0] != '"' || unquoted[len(unquoted)-1] != '"' {
		return errors.New("byte: invalid byte " + unquoted)
	}
	unquoted = unquoted[1 : len(unquoted)-1]
	suffixPos := strings.IndexFunc(unquoted, func(c rune) bool {
		return c != '.' && !unicode.IsDigit(c)
	})
	if suffixPos == -1 {
		suffixPos = len(unquoted)
	}
	parsed, err := strconv.ParseFloat(unquoted[:suffixPos], 64)
	if err != nil {
		return err
	}
	unit := Byte(1)
	switch string(unquoted[suffixPos:]) {
	case "TiB":
		unit = Tebibyte
	case "GiB":
		unit = Gibibyte
	case "MiB":
		unit = Mebibyte
	case "KiB":
		unit = Kibibyte
	case "B":
	case "":
		unit = Byte(1)
	default:
		return errors.New("byte: invalid byte " + string(data))
	}
	*b = Byte(parsed * float64(unit))
	return nil
}
