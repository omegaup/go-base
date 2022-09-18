package base

import (
	"testing"
)

func TestByte(t *testing.T) {
	testTable := []struct {
		str      string
		expected Byte
	}{
		{"1", Byte(1)},
		{"\"10\"", Byte(10)},
		{"\"100B\"", Byte(100)},
		{"\"0.5KiB\"", Byte(512)},
		{"\"1KiB\"", Kibibyte},
		{"\"1MiB\"", Mebibyte},
		{"\"1GiB\"", Gibibyte},
		{"\"1TiB\"", Tebibyte},
	}
	for _, entry := range testTable {
		var b Byte
		if err := b.UnmarshalJSON([]byte(entry.str)); err != nil {
			t.Fatalf(err.Error())
		}
		if entry.expected != b {
			t.Errorf("expected %v got %v", entry.expected, b)
		}
		marshaled, err := b.MarshalJSON()
		if err != nil {
			t.Fatalf(err.Error())
		}
		var b2 Byte
		if err := b2.UnmarshalJSON(marshaled); err != nil {
			t.Fatalf(err.Error())
		}
		if entry.expected != b2 {
			t.Errorf("expected %v got %v", entry.expected, b2)
		}
	}
}

func TestByteNil(t *testing.T) {
	var b, zeroValue Byte
	if err := b.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatalf(err.Error())
	}
	if zeroValue != b {
		t.Errorf("expected %v got %v", zeroValue, b)
	}
}

func TestByteNumber(t *testing.T) {
	var b Byte
	if err := b.UnmarshalJSON([]byte("1099511627776")); err != nil {
		t.Fatalf(err.Error())
	}
	if 1099511627776 != b.Bytes() {
		t.Errorf("expected 1099511627776 got %v", b)
	}
	if 1073741824 != b.Kibibytes() {
		t.Errorf("expected 1073741824 got %v", b)
	}
	if 1048576 != b.Mebibytes() {
		t.Errorf("expected 1048576 got %v", b)
	}
	if 1024 != b.Gibibytes() {
		t.Errorf("expected 1024 got %v", b)
	}
	if 1 != b.Tebibytes() {
		t.Errorf("expected 1 got %v", b)
	}
}

func TestByteInvalidNumber(t *testing.T) {
	var b, zeroValue Byte
	if err := b.UnmarshalJSON([]byte("1.x")); err != nil {
		t.Fatalf(err.Error())
	}
	if zeroValue != b {
		t.Errorf("expected %v got %v", zeroValue, b)
	}
}

func TestByteInvalidString(t *testing.T) {
	var b Byte
	if err := b.UnmarshalJSON([]byte("\"1")); err == nil {
		t.Errorf("Expected an error, but got nil")
	}
	if err := b.UnmarshalJSON([]byte("\"1.x\"")); err == nil {
		t.Errorf("Expected an error, but got nil")
	}
}

func TestByteComparison(t *testing.T) {
	b1 := Byte(1)
	b2 := Byte(2)
	if b1 != Min(b1, b2) {
		t.Errorf("expected %v got %v", b1, Min(b1, b2))
	}
	if b1 != Min(b2, b1) {
		t.Errorf("expected %v got %v", b1, Min(b2, b1))
	}
	if b2 != Max(b1, b2) {
		t.Errorf("expected %v got %v", b2, Max(b1, b2))
	}
	if b2 != Max(b2, b1) {
		t.Errorf("expected %v got %v", b2, Max(b2, b1))
	}
}
