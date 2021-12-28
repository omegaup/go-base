package base

import (
	"fmt"
	"math/big"
	"testing"
)

func TestParseRational(t *testing.T) {
	testTable := []struct {
		str      string
		expected *big.Rat
	}{
		{"1", big.NewRat(1, 1)},
		{"0.5", big.NewRat(1, 2)},
		{"0.333333333", big.NewRat(1, 3)},
		{"0.23", big.NewRat(23, 100)},
		{"0.023", big.NewRat(23, 1000)},
		{"0.208333333", big.NewRat(5, 24)},
		{"0.123456789", big.NewRat(63, 512)},
	}
	for _, entry := range testTable {
		entry := entry
		t.Run(entry.str, func(t *testing.T) {
			var val *big.Rat
			var err error
			if val, err = ParseRational(entry.str); err != nil {
				t.Fatalf(err.Error())
			}
			if entry.expected.Cmp(val) != 0 {
				t.Errorf("expected %v got %v", entry.expected, val)
			}
		})
	}
}

func TestRatMarshalRoundtripString(t *testing.T) {
	for _, r1 := range []*Rat{
		(*Rat)(big.NewRat(1, 3)),
		(*Rat)(big.NewRat(-1, 1)),
		(*Rat)(big.NewRat(1, 1)),
		(*Rat)(big.NewRat((int64(1)<<int64(53))-int64(1), 1)),
		(*Rat)(big.NewRat(0, 1)),
	} {
		r1 := r1
		t.Run(fmt.Sprintf("%v", r1), func(t *testing.T) {
			serialized, err := r1.MarshalJSON()
			if err != nil {
				t.Fatalf(err.Error())
			}
			var r2 Rat
			if err = r2.UnmarshalJSON(serialized); err != nil {
				t.Fatalf(err.Error())
			}
			if (*big.Rat)(r1).Cmp((*big.Rat)(&r2)) != 0 {
				t.Errorf("expected %v got %v", (*big.Rat)(r1).String(), (*big.Rat)(&r2).String())
			}
		})
	}
}

func TestRatUnmarshalNumber(t *testing.T) {
	for s, expected := range map[string]*Rat{
		"\"1/3\"":              (*Rat)(big.NewRat(1, 3)),
		"-1":                   (*Rat)(big.NewRat(-1, 1)),
		"1":                    (*Rat)(big.NewRat(1, 1)),
		"4503599627370495":     (*Rat)(big.NewRat((int64(1)<<int64(52))-int64(1), 1)),
		"4503599627370496":     (*Rat)(big.NewRat((int64(1) << int64(52)), 1)),
		"9007199254740991":     (*Rat)(big.NewRat((int64(1)<<int64(53))-int64(1), 1)),
		"\"9007199254740992\"": (*Rat)(big.NewRat((int64(1) << int64(53)), 1)),
		"0":                    (*Rat)(big.NewRat(0, 1)),
	} {
		t.Run(fmt.Sprintf("unmarshal %s", s), func(t *testing.T) {
			var r Rat
			err := r.UnmarshalJSON([]byte(s))
			if err != nil {
				t.Fatalf(err.Error())
			}
			if (*big.Rat)(expected).Cmp((*big.Rat)(&r)) != 0 {
				t.Errorf("expected %v got %v", (*big.Rat)(expected).String(), (*big.Rat)(&r).String())
			}
		})
		t.Run(fmt.Sprintf("marshal %s", s), func(t *testing.T) {
			serialized, err := expected.MarshalJSON()
			if err != nil {
				t.Fatalf(err.Error())
			}
			if s != string(serialized) {
				t.Errorf("expected %v got %v", s, string(serialized))
			}
		})
	}
}
