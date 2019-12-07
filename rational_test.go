package base

import (
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
		var val *big.Rat
		var err error
		if val, err = ParseRational(entry.str); err != nil {
			t.Fatalf(err.Error())
		}
		if entry.expected.Cmp(val) != 0 {
			t.Errorf("expected %v got %v", entry.expected, val)
		}
	}
}
